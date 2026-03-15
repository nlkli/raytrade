package core

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"nlkli/raytrade/internal/broker"
	"nlkli/raytrade/internal/cdl"
	"nlkli/raytrade/internal/utils"
	"nlkli/raytrade/internal/ws"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
	"golang.org/x/exp/slices"
)

const (
	REQUEST_TIMEOUT         = time.Second * 3
	MAX_CATEGORY_STREAMS    = 4
	MAX_TOPIC_SUBSCRIPTIONS = 24
)

type Task interface {
	Execute(*Background) error
}

type SelectInstrumentPrice struct {
	Category broker.Category
	Symbol   string
	Price    float64
}

func (t *SelectInstrumentPrice) Execute(b *Background) error {
	b.GetOrInitInstrumentInfo(t.Category, t.Symbol).SelectedPrice.Store(&t.Price)
	return nil
}

type SelectOrder struct {
	Order *broker.Order
}

func (t *SelectOrder) Execute(b *Background) error {
	b.selectedOrder.Store(t.Order)
	return nil
}

type SelectPosition struct {
	Position *broker.Position
}

func (t *SelectPosition) Execute(b *Background) error {
	b.selectedPosition.Store(t.Position)
	return nil
}

type OrderBy int

const (
	OrderIndex OrderBy = iota
	SelectedOrder
	FirstOrder
	LastOrder
)

type CancelOrder struct {
	OrderBy      OrderBy
	OrderByValue any
}

func (t *CancelOrder) Execute(b *Background) error {
	var co *broker.Order
	switch t.OrderBy {
	case OrderIndex:
		idx, ok := t.OrderByValue.(int)
		if !ok {
			return fmt.Errorf("order index must be int")
		}

		orderPtr := b.order.Load()
		if orderPtr == nil {
			return fmt.Errorf("no orders")
		}

		order := *orderPtr

		if idx < 0 || idx >= len(order) {
			return fmt.Errorf("order index out of range")
		}
		co = &order[idx]

	case SelectedOrder:
		so := b.selectedOrder.Load()

		if so == nil {
			return fmt.Errorf("no selected order")
		}
		co = so

	case FirstOrder:
		orderPtr := b.order.Load()
		if orderPtr == nil {
			return fmt.Errorf("no orders")
		}

		order := *orderPtr

		if len(order) == 0 {
			return fmt.Errorf("no orders")
		}
		co = &order[0]

	case LastOrder:
		orderPtr := b.order.Load()
		if orderPtr == nil {
			return fmt.Errorf("no orders")
		}

		order := *orderPtr

		n := len(order)
		if n == 0 {
			return fmt.Errorf("no orders")
		}
		co = &order[n-1]

	}

	ctx, cancel := context.WithTimeout(
		context.Background(),
		REQUEST_TIMEOUT,
	)
	_, _, err := b.broker.CancelOrder(
		ctx, co.Category, co.Symbol, co.Id, "",
	)
	cancel()

	return err
}

type OrderPriceBy int

const (
	Price OrderPriceBy = iota
	SelectedPrice
	BidPrice
	AskPrice
)

type PlaceLimitOrder struct {
	Category broker.Category
	Symbol   string

	Side broker.Side

	Qty        float64
	MarketUnit broker.MarketUnit

	PriceBy      OrderPriceBy
	PriceByValue any
}

func (t *PlaceLimitOrder) Execute(b *Background) error {
	params := broker.PlaceOrderParams{
		Category:   t.Category,
		Symbol:     t.Symbol,
		Side:       t.Side,
		Type:       broker.Limit,
		Qty:        t.Qty,
		MarketUnit: &t.MarketUnit,
	}

	var price float64
	switch t.PriceBy {
	case Price:
		p, ok := t.PriceByValue.(float64)
		if !ok {
			return fmt.Errorf("price must be float64")
		}
		price = p

	case SelectedPrice:
		instrument := b.GetInstrumentInfo(t.Category, t.Symbol)
		if instrument == nil {
			return fmt.Errorf("instrument not found")
		}
		sp := instrument.SelectedPrice.Load()
		if sp == nil {
			return fmt.Errorf("select price not set")
		}
		price = *sp

	case BidPrice, AskPrice:
		instrument := b.GetInstrumentInfo(t.Category, t.Symbol)
		if instrument == nil {
			return fmt.Errorf("instrument not found")
		}

		ob := instrument.OrderBook.Load()
		if ob == nil {
			return fmt.Errorf("order book not available")
		}

		sideIndex := 0 // bids
		if t.PriceBy == AskPrice {
			sideIndex = 1 // asks
		}

		entries := (*ob)[sideIndex]

		idx, ok := t.PriceByValue.(int)
		if !ok {
			return fmt.Errorf("price index must be int")
		}

		if idx < 0 || idx >= len(entries) {
			return fmt.Errorf("price index out of range")
		}

		price = entries[idx][0]
	}

	params.Price = &price

	ctx, cancel := context.WithTimeout(context.Background(), REQUEST_TIMEOUT)
	_, _, err := b.broker.PlaceOrder(ctx, "", params)
	cancel()

	return err
}

type ExtendStartCandles struct {
	Idx      int
	Category broker.Category
	Symbol   string
	Interval cdl.Interval
	Candles  []cdl.Candle
	Limit    int
}

func (t *ExtendStartCandles) Execute(b *Background) error {
	defer b.Push(func(s *State) {
		cs := s.Chart[t.Idx]

		if cs == nil {
			return
		}

		cs.ExtendCandlesF = false
	})

	if len(t.Candles) == 0 {
		return errors.New("empty candles")
	}

	if t.Limit <= 1 {
		return errors.New("ivalid limit param")
	}

	startCandles := len(t.Candles)
	exCandles := t.Candles
	// intervalMilli := int64(t.Interval.AsMilli())
	timeout := time.After(time.Minute)

loop:
	for {
		select {
		case <-timeout:
			break loop
		default:
			ctx, cancel := context.WithTimeout(
				context.Background(),
				REQUEST_TIMEOUT,
			)
			res, err := b.broker.ExtendStartCandles(
				ctx,
				exCandles,
				t.Category,
				t.Symbol,
				t.Interval,
				t.Limit,
			)

			cancel()

			if err == nil {

				exCandles = append(res, exCandles...)

				b.Push(func(s *State) {
					cs := s.Chart[t.Idx]

					if cs == nil {
						return
					}

					if cs.ExtendCandlesF {
						cs.Candles = exCandles[:len(exCandles)-startCandles]
						startCandles = len(cs.Candles)

						cs.Forced = true
					}
				})

				if len(res) < t.Limit {
					continue
				}

				break loop
			}
		}
	}

	return nil
}

type SubChart struct {
	Idx      int
	Category broker.Category
	Symbol   string
	Interval cdl.Interval
}

func (t *SubChart) Execute(b *Background) error {
	stream := b.GetOrInitStream(t.Category)

	if b.chartSub[t.Idx] != nil {
		b.chartSub[t.Idx].Stop()
	}

	sub, err := stream.SubscribeCandle(t.Symbol, t.Interval)
	if err != nil {
		return fmt.Errorf(
			"candle subscribtion %s.%s.%s error: %e",
			t.Category.AsString(false),
			t.Symbol,
			t.Interval.AsString(),
			err,
		)
	}

	b.chartSub[t.Idx] = sub

	first := <-sub.C

	b.Push(func(s *State) {
		cs := s.Chart[t.Idx]

		if cs == nil {
			return
		}

		cString := t.Category.AsString(true)
		iString := t.Interval.AsString()

		cs.InstrumentKey = fmt.Sprintf(
			"%s.%s", cString, t.Symbol,
		)
		cs.Category = t.Category
		cs.Symbol = t.Symbol
		cs.Interval = t.Interval

		pi := slices.IndexFunc(s.Position.List, func(e broker.Position) bool {
			return cs.InstrumentKey == fmt.Sprintf(
				"%s.%s",
				e.Category.AsString(true),
				e.Symbol,
			)
		})

		cs.PositionIdx = pi

		cs.LableString = fmt.Sprintf(
			"%d.%s.%s.%s", t.Idx, cString, cs.Symbol, iString,
		)

		cs.SecInterval = float32(t.Interval.AsSeconds())
		cs.Candles = []cdl.Candle{first.Candle}
		cs.ExtendCandlesF = false

		cs.Levels = nil
		cs.Lines = nil
		cs.IsLineDuring = false
		cs.Forced = true
	})

	go func() {
		intervalMilli := int64(t.Interval.AsMilli())

		for d := range sub.C {
			if d.Confirm {
				b.Push(func(s *State) {
					cs := s.Chart[t.Idx]

					if cs == nil {
						return
					}

					cs.LastPrice = d.Candle.C

					n := len(cs.Candles)

					timeDiff := d.Candle.Time - cs.Candles[n-2].Time
					if timeDiff > intervalMilli+200 {
						cs.Candles = []cdl.Candle{d.Candle}
					} else {
						cs.Candles = append(cs.Candles, d.Candle)
					}

					cs.Forced = true
				})
				continue
			}

			b.Push(func(s *State) {
				cs := s.Chart[t.Idx]

				if cs == nil {
					return
				}

				cs.LastPrice = d.Candle.C
				cs.Candles[len(cs.Candles)-1] = d.Candle

				cs.MaxP = max(cs.MaxP, d.Candle.H)
				cs.MinP = min(cs.MinP, d.Candle.L)
				cs.MidP = (cs.MaxP + cs.MinP) * .5
				cs.RngP = cs.MaxP - cs.MinP
			})
		}
	}()

	return nil
}

type SubOrderBook struct {
	Idx      int
	Category broker.Category
	Symbol   string
	Depth    int
}

func (t *SubOrderBook) Execute(b *Background) error {

	stream := b.GetOrInitStream(t.Category)

	if b.orderBookSub[t.Idx] != nil {
		b.orderBookSub[t.Idx].Stop()
	}

	sub, err := stream.SubscribeOrderBook(t.Symbol, t.Depth)
	if err != nil {
		return fmt.Errorf(
			"orderbook subscribtion error: %e",
			err,
		)
	}

	instrumentKey := fmt.Sprintf(
		"%s.%s", t.Category.AsString(true), t.Symbol,
	)

	b.Push(func(s *State) {
		obS := s.OrderBook[t.Idx]

		if obS == nil {
			return
		}

		obS.InstrumentKey = instrumentKey
		obS.Category = t.Category
		obS.Symbol = t.Symbol

		obS.Forced = true
	})

	b.orderBookSub[t.Idx] = sub

	go func() {
		instrument := b.GetOrInitInstrumentInfo(t.Category, t.Symbol)

		for d := range sub.C {
			instrument.OrderBook.Store(&d)

			b.Push(func(s *State) {
				obS := s.OrderBook[t.Idx]

				if obS == nil {
					return
				}

				obS.Bids = d[0]
				obS.Asks = d[1]

				obS.Forced = true
			})
		}
	}()

	return nil
}

type InstrumentFilter struct {
	Category broker.Category
	Symbol   string
}

type SubExecution struct {
	Filter []InstrumentFilter
	Limit  int
}

func (t *SubExecution) Execute(b *Background) error {

	stream := b.GetOrInitPrivateStream()

	b.executionMu.Lock()

	if b.executionSub != nil {
		b.executionSub.Stop()
	}

	sub, err := stream.SubscribeExecution()
	if err != nil {
		return err
	}

	b.executionSub = sub

	b.executionMu.Unlock()

	go func() {
		for d := range b.executionSub.C {
			if len(t.Filter) > 0 {
				if !slices.ContainsFunc(t.Filter, func(e InstrumentFilter) bool {
					return e.Category == d.Category && e.Symbol == d.Symbol
				}) {
					continue
				}
			}
			b.Push(func(s *State) {
				if len(s.Execution.List) > t.Limit*2 {
					s.Execution.List = append(
						[]broker.Execution{},
						s.Execution.List[t.Limit:]...,
					)
				}
				s.Execution.List = append(s.Execution.List, d)
				s.Execution.Forced = true
			})
		}
	}()

	return nil
}

type SubPosition struct {
	Filter []InstrumentFilter
}

func (t *SubPosition) Execute(b *Background) error {

	stream := b.GetOrInitPrivateStream()

	b.positionMu.Lock()

	if b.positionSub != nil {
		b.positionSub.Stop()
	}

	sub, err := stream.SubscribePosition()
	if err != nil {
		return err
	}

	b.positionSub = sub

	b.positionMu.Unlock()

	updatePosition := func(p broker.Position) {
		b.GetOrInitInstrumentInfo(p.Category, p.Symbol).Position.Store(&p)

		instrumentKey := fmt.Sprintf(
			"%s.%s",
			p.Category.AsString(true),
			p.Symbol,
		)

		b.Push(func(s *State) {
			s.Position.Forced = true

			i := slices.IndexFunc(
				s.Position.List,
				func(e broker.Position) bool {
					return p.Symbol == e.Symbol &&
						p.Category == e.Category
				},
			)

			if i >= 0 {
				s.Position.List[i] = p
				return
			}

			s.Position.List = append(s.Position.List, p)

			for _, ch := range s.Chart {
				if ch.InstrumentKey == instrumentKey {
					ch.PositionIdx = len(s.Position.List) - 1
				}
			}
		})
	}

	done := make(chan struct{}, 1)
	go func() {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()

		for {
			for _, f := range t.Filter {
				ctx, cancel := context.WithTimeout(
					context.Background(),
					REQUEST_TIMEOUT,
				)
				res, err := b.broker.GetPosition(
					ctx, f.Category, f.Symbol,
				)
				cancel()

				if err != nil {
					continue
				}

				for _, p := range res {
					updatePosition(p)
				}
			}

			select {
			case <-ticker.C:
			case <-done:
				return
			}
		}

	}()

	go func() {
		defer close(done)

		for d := range b.positionSub.C {
			if len(t.Filter) > 0 {
				if !slices.ContainsFunc(t.Filter, func(e InstrumentFilter) bool {
					return e.Category == d.Category && e.Symbol == d.Symbol
				}) {
					continue
				}
			}
			updatePosition(d)
		}
	}()

	return nil
}

type SubOrder struct {
	Idx    int
	Filter []InstrumentFilter
}

func (t *SubOrder) Execute(b *Background) error {
	stream := b.GetOrInitPrivateStream()

	b.orderMu.Lock()

	if b.orderSub != nil {
		b.orderSub.Stop()
	}

	sub, err := stream.SubscribeOrder()
	if err != nil {
		return err
	}

	b.orderSub = sub

	b.orderMu.Unlock()

	done := make(chan struct{}, 1)
	go func() {
		ticker := time.NewTicker(time.Second * 15)
		defer ticker.Stop()

	loop:
		for {
			var orderList []broker.Order

			for _, f := range t.Filter {
				ctx, cancel := context.WithTimeout(
					context.Background(),
					REQUEST_TIMEOUT,
				)
				// TODO nextPageCursor
				res, _, err := b.broker.GetOpenOrder(
					ctx, f.Category, f.Symbol,
				)
				cancel()

				if err != nil {
					continue loop
				}

				orderList = append(orderList, res...)
			}

			b.Push(func(s *State) {
				s.Order.List = orderList
				b.order.Store(&s.Order.List)
				s.Order.Forced = true
			})

			select {
			case <-ticker.C:
			case <-done:
				return
			}
		}

	}()

	go func() {
		defer close(done)

		for o := range sub.C {
			if len(t.Filter) > 0 {
				if !slices.ContainsFunc(t.Filter, func(e InstrumentFilter) bool {
					return e.Category == o.Category && e.Symbol == o.Symbol
				}) {
					continue
				}
			}

			b.Push(func(s *State) {
				s.Order.Forced = true
				i := slices.IndexFunc(
					s.Order.List,
					func(e broker.Order) bool {
						return o.Id == e.Id
					},
				)
				if i >= 0 {
					if o.Status == broker.Closed || o.LeavesQty == 0 {
						s.Order.List = append(
							s.Order.List[:i],
							s.Order.List[i+1:]...,
						)
						b.order.Store(&s.Order.List)
						return
					}
					s.Order.List[i] = o
					return
				}
				if o.Status != broker.Closed {
					s.Order.List = append([]broker.Order{o}, s.Order.List...)
					b.order.Store(&s.Order.List)
				}
			})
		}
	}()

	return nil
}

type InstrumentInfo struct {
	Key      string
	Category broker.Category
	Symbol   string

	Position      atomic.Pointer[broker.Position]
	OrderBook     atomic.Pointer[[2][][2]float64]
	SelectedPrice atomic.Pointer[float64]
}

type Background struct {
	Tx chan Task

	commit atomic.Pointer[CommitFn]

	broker broker.Broker

	stream   [MAX_CATEGORY_STREAMS]broker.Stream // public streams
	streamMu sync.Mutex

	privateStream   broker.PrivateStream
	privateStreamMu sync.Mutex

	executionSub *broker.Subscription[broker.Execution]
	executionMu  sync.Mutex

	positionSub *broker.Subscription[broker.Position]
	positionMu  sync.Mutex

	orderSub *broker.Subscription[broker.Order]
	orderMu  sync.Mutex

	chartSub     [MAX_TOPIC_SUBSCRIPTIONS]*broker.Subscription[cdl.CandleStreamData]
	orderBookSub [MAX_TOPIC_SUBSCRIPTIONS]*broker.Subscription[[2][][2]float64]

	instrument   []*InstrumentInfo
	instrumentMu sync.Mutex

	order         atomic.Pointer[[]broker.Order]
	selectedOrder atomic.Pointer[broker.Order]

	selectedPosition atomic.Pointer[broker.Position]
}

func InitBackground(

	ctx context.Context,
	b broker.Broker,

) *Background {
	bg := &Background{
		Tx:     make(chan Task, 1),
		broker: b,
	}

	const totalWorkers = 8
	workers := [totalWorkers]chan Task{}
	for i := range workers {
		ch := make(chan Task)
		workers[i] = ch
		go func() {
			for t := range ch {
				if err := t.Execute(bg); err != nil {
					bg.Push(
						CommitCommandLineError(err.Error()),
					)
				}
			}
		}()
	}

	go func() {
		wi := 0

		for {
			select {
			case t := <-bg.Tx:
				workers[wi] <- t
				wi++
				if wi >= totalWorkers {
					wi = 0
				}
			case <-ctx.Done():
				if bg.privateStream != nil {
					bg.privateStream.Close()
				}
				for _, s := range bg.stream {
					if s == nil {
						continue
					}
					s.Close()
				}
				for _, ch := range workers {
					close(ch)
				}
				return
			}
		}
	}()

	return bg
}

func (b *Background) GetInstrumentInfo(c broker.Category, symbol string) *InstrumentInfo {
	b.instrumentMu.Lock()
	defer b.instrumentMu.Unlock()

	ii := slices.IndexFunc(b.instrument, func(e *InstrumentInfo) bool {
		return e.Category == c && e.Symbol == symbol
	})

	if ii < 0 {
		return nil
	}

	return b.instrument[ii]
}

func (b *Background) GetOrInitInstrumentInfo(c broker.Category, symbol string) *InstrumentInfo {
	b.instrumentMu.Lock()
	defer b.instrumentMu.Unlock()

	ii := slices.IndexFunc(b.instrument, func(e *InstrumentInfo) bool {
		return e.Category == c && e.Symbol == symbol
	})

	if ii < 0 {
		instrument := &InstrumentInfo{
			Key:      fmt.Sprintf("%s.%s", c.AsString(true), symbol),
			Category: c,
			Symbol:   symbol,
		}
		b.instrument = append(b.instrument, instrument)
		return instrument
	}

	return b.instrument[ii]
}

func (b *Background) WatchConfig(ctx context.Context, configPath string) error {

	watchConfig, err := utils.WatchFile(ctx, configPath, time.Second*3)
	if err != nil {
		return err
	}

	go func() {
		for {
			select {
			case cb := <-watchConfig:
				var c Config
				err := json.Unmarshal(cb, &c)
				if err != nil {
					CommitCommandLineError(err.Error())
					continue
				}
				b.Push(func(s *State) {
					s.ApplyConfig(&c)
				})
			case <-ctx.Done():
				return
			}
		}
	}()

	return nil
}

func (b *Background) GetOrInitPrivateStream() broker.PrivateStream {
	b.privateStreamMu.Lock()
	defer b.privateStreamMu.Unlock()

	if b.privateStream == nil {
		b.privateStream = b.broker.CreatePrivateStream(
			func(conn *websocket.Conn, _ chan<- []byte, _ int) error {
				return nil
			},
		)
		time.Sleep(time.Second * 5)
	}

	return b.privateStream
}

func (b *Background) GetOrInitStream(c broker.Category) broker.Stream {
	b.streamMu.Lock()
	defer b.streamMu.Unlock()

	if b.stream[c] == nil {
		b.stream[c] = b.broker.CreateStream(
			c,
			func(conn *websocket.Conn, _ chan<- []byte, _ int) error {
				return nil
			},
			ws.WithOnConnectFn(func(i int) {
				b.Push(func(s *State) {
					s.CommandLine.Prompt = "stream connection established"
					s.CommandLine.Color = s.P.Dim.Green
				})
			}),
			ws.WithOnDisconnectFn(
				func(_ int, _ bool) {
					b.Push(func(s *State) {
						s.CommandLine.Prompt = "stream lost. Reconnecting..."
						s.CommandLine.Color = s.P.Dim.Yellow
					})
				},
			),
		)
	}

	return b.stream[c]
}

func (b *Background) Push(f CommitFn) {
	for {
		current := b.commit.Load()

		var newFn CommitFn
		if current == nil {
			newFn = f
		} else {
			oldFn := *current
			newFn = func(s *State) {
				oldFn(s)
				f(s)
			}
		}

		if b.commit.CompareAndSwap(current, &newFn) {
			return
		}
	}
}

func (b *Background) Update(s *State) {
	if fn := b.commit.Swap(nil); fn != nil {
		(*fn)(s)
	}
}

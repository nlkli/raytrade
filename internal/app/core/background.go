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

	"golang.org/x/exp/slices"
)

const (
	REQUEST_TIMEOUT         = time.Second * 3
	MAX_CATEGORY_STREAMS    = 4
	MAX_TOPIC_SUBSCRIPTIONS = 16
)

type Task interface {
	Execute(*Background) error
}

type PlaceOrder struct {
	Category     broker.Category
	Symbol       string
	UseMargin    bool
	Side         broker.Side
	Type         broker.OrderType
	Qty          float64
	IsCoinQty    bool
	Price        float64
	TriggerPrice float64
	TriggerBy    string
	TakeProfit   float64
	StopLoss     float64
}

func (t *PlaceOrder) Execute(b *Background) error {
	return nil
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
		s.Chart[t.Idx].ExtendCandlesF = false
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

	if b.chartSub[t.Idx] != nil {
		if err = b.chartSub[t.Idx].Stop(); err != nil {
			sub.Stop()
			return fmt.Errorf(
				"candle unsubscribtion error: %e",
				err,
			)
		}
	}

	b.chartSub[t.Idx] = sub

	first := <-sub.C

	b.Push(func(s *State) {
		cs := s.Chart[t.Idx]

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
			"%s.%s.%s", cString, cs.Symbol, iString,
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

	sub, err := stream.SubscribeOrderBook(t.Symbol, t.Depth)
	if err != nil {
		return fmt.Errorf(
			"orderbook subscribtion error: %e",
			err,
		)
	}

	b.Push(func(s *State) {
		obS := s.OrderBook[t.Idx]

		obS.InstrumentKey = fmt.Sprintf(
			"%s.%s", t.Category.AsString(true), t.Symbol,
		)
		obS.Category = t.Category
		obS.Symbol = t.Symbol

		obS.Forced = true
	})

	if b.orderBookSub[t.Idx] != nil {
		b.orderBookSub[t.Idx].Stop()
	}
	b.orderBookSub[t.Idx] = sub

	go func() {
		for d := range sub.C {
			b.Push(func(s *State) {
				obS := s.OrderBook[t.Idx]

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

type SubPosition struct {
	Filter []InstrumentFilter
}

func (t *SubPosition) Execute(b *Background) error {

	stream := b.GetOrInitPrivateStream()

	sub, err := stream.SubscribePosition()
	if err != nil {
		return err
	}

	b.positionMu.Lock()
	if b.positionSub != nil {
		b.positionSub.Stop()
	}
	b.positionSub = sub
	b.positionMu.Unlock()

	updatePosition := func(p broker.Position) {
		b.Push(func(s *State) {
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
			instrumentKey := fmt.Sprintf(
				"%s.%s",
				p.Category.AsString(true),
				p.Symbol,
			)
			si := 0
			for {
				ci := slices.IndexFunc(s.Chart[si:], func(e *ChartState) bool {
					return e.InstrumentKey == instrumentKey
				})
				if ci >= 0 {
					s.Chart[ci+si].PositionIdx = len(s.Position.List) - 1
					si = ci + 1
					if si >= len(s.Chart) {
						break
					}
					continue
				}
				break
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

	sub, err := stream.SubscribeOrder()
	if err != nil {
		return err
	}

	b.orderMu.Lock()
	if b.orderSub != nil {
		b.orderSub.Stop()
	}
	b.orderSub = sub
	b.orderMu.Unlock()

	updateOrder := func(o broker.Order) {
		b.Push(func(s *State) {
			i := slices.IndexFunc(
				s.Order.List,
				func(e broker.Order) bool {
					return o.Id == e.Id
				},
			)
			if i >= 0 {
				if o.Status == broker.Closed {
					s.Order.List = slices.Delete(s.Order.List, i, i+1)
					return
				}
				s.Order.List[i] = o
				return
			}
			s.Order.List = append(s.Order.List, o)
		})
	}

	done := make(chan struct{}, 1)
	go func() {
		ticker := time.NewTicker(time.Second * 15)
		defer ticker.Stop()

		for {
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
					continue
				}

				for _, o := range res {
					updateOrder(o)
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

		for o := range sub.C {
			if len(t.Filter) > 0 {
				if !slices.ContainsFunc(t.Filter, func(e InstrumentFilter) bool {
					return e.Category == o.Category && e.Symbol == o.Symbol
				}) {
					continue
				}
			}
			updateOrder(o)
		}
	}()

	return nil
}

type Background struct {
	Tx chan Task

	commit atomic.Pointer[CommitFn]

	broker broker.Broker

	stream   [MAX_CATEGORY_STREAMS]broker.Stream // public streams
	streamMu sync.Mutex

	privateStream   broker.PrivateStream
	privateStreamMu sync.Mutex

	positionSub *broker.Subscription[broker.Position]
	positionMu  sync.Mutex

	orderSub *broker.Subscription[broker.Order]
	orderMu  sync.Mutex

	chartSub     [MAX_TOPIC_SUBSCRIPTIONS]*broker.Subscription[cdl.CandleStreamData]
	orderBookSub [MAX_TOPIC_SUBSCRIPTIONS]*broker.Subscription[[2][][2]float64]
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
		b.privateStream = b.broker.CreatePrivateStream(nil)
		time.Sleep(time.Second * 5)
	}

	return b.privateStream
}

func (b *Background) GetOrInitStream(c broker.Category) broker.Stream {
	b.streamMu.Lock()
	defer b.streamMu.Unlock()

	if b.stream[c] == nil {
		b.stream[c] = b.broker.CreateStream(
			c, nil,
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

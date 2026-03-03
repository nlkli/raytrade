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

type PositionFilter struct {
	Category broker.Category
	Symbol   string
}

type SubPosition struct {
	Filter []PositionFilter
}

func (t *SubPosition) Execute(b *Background) error {

	stream := b.GetOrInitPrivateStream()

	ctx, cancel := context.WithTimeout(context.Background(), REQUEST_TIMEOUT)
	defer cancel()

	position, err := b.broker.GetPosition(ctx, broker.Futures)
	if err != nil {
		println(err.Error())
		return err
	}

	b.Push(func(s *State) {
		s.Position.List = position
	})

	if b.positionSub != nil {
		b.positionSub.Stop()
	}
	b.positionSub, err = stream.SubscribePosition()

	if err != nil {
		println(err.Error())
		return err
	}

	go func() {
		for d := range b.positionSub.C {
			if len(t.Filter) > 0 {
				if !slices.ContainsFunc(t.Filter, func(e PositionFilter) bool {
					return e.Category == d.Category && e.Symbol == d.Symbol
				}) {
					continue
				}
			}
			b.Push(func(s *State) {
				i := slices.IndexFunc(s.Position.List, func(e broker.Position) bool {
					return e.CreatedAt.Equal(d.CreatedAt)
				})
				if i >= 0 {
					s.Position.List[i] = d
					return
				}
				s.Position.List = append(s.Position.List, d)
			})
		}
	}()

	return nil
}

type OrderFilter struct {
	Category broker.Category
	Symbol   string
}

type SubOrder struct {
	Idx    int
	Filter []PositionFilter
}

func (t *SubOrder) Execute(b *Background) error {
	// stream := b.GetOrInitPrivateStream()

	// sub, err := stream.SubscribeOrder()

	// if err != nil {
	// 	return err
	// }

	go func() {
		// for _ := range sub.C {
		// }
	}()

	return nil
}

type Background struct {
	Tx chan Task

	commit atomic.Pointer[CommitFn]

	broker broker.Broker

	stream        [MAX_CATEGORY_STREAMS]broker.Stream // public streams
	privateStream broker.PrivateStream

	positionSub *broker.Subscription[broker.Position]
	orderSub    *broker.Subscription[broker.Order]

	chartSub     [MAX_TOPIC_SUBSCRIPTIONS]*broker.Subscription[cdl.CandleStreamData]
	orderBookSub [MAX_TOPIC_SUBSCRIPTIONS]*broker.Subscription[[2][][2]float64]

	mu sync.Mutex
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
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.privateStream == nil {
		b.privateStream = b.broker.CreatePrivateStream(nil)
	}

	return b.privateStream
}

func (b *Background) GetOrInitStream(c broker.Category) broker.Stream {
	b.mu.Lock()
	defer b.mu.Unlock()

	stream := b.stream[c]

	if stream == nil {
		stream = b.broker.CreateStream(
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
						s.CommandLine.Prompt = "stream Lost. Reconnecting..."
						s.CommandLine.Color = s.P.Dim.Yellow
					})
				},
			),
		)
		b.stream[c] = stream
		return stream
	}

	return stream
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

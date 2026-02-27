package core

import (
	"context"
	"fmt"
	"nlkli/raytrade/internal/broker"
	"nlkli/raytrade/internal/cdl"
	"nlkli/raytrade/internal/ws"
	"sync/atomic"
	"time"
)

const (
	MAX_CATEGORY_STREAMS    = 4
	MAX_TOPIC_SUBSCRIPTIONS = 16
)

type Task interface {
	Execute(*Background) error
}

type PlaceOrder struct {
	Category broker.Category
	Symbol   string
	Side     broker.Side
	Price    float64
	UsdQty   float64
}

func (t *PlaceOrder) Execute(b *Background) error {
	return nil
}

type SubChart struct {
	Idx      int
	Category broker.Category
	Symbol   string
	Interval cdl.Interval
	Limit    int
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

	first := <-sub.C

	if b.chartSub[t.Idx] != nil {
		if err = b.chartSub[t.Idx].Stop(); err != nil {
			return fmt.Errorf(
				"candle unsubscribtion error: %e",
				err,
			)
		}
	}

	b.chartSub[t.Idx] = sub

	b.Push(func(s *State) {
		cs := s.Chart[t.Idx]

		cs.Category = t.Category.AsString(true)
		cs.Symbol = t.Symbol
		cs.Interval = t.Interval.AsString()

		cs.LableString = fmt.Sprintf(
			"%s.%s.%s", cs.Category, cs.Symbol, cs.Interval,
		)

		cs.SecInterval = float32(t.Interval.AsSeconds())
		cs.Candles = []cdl.Candle{first.Candle}
		cs.Levels = nil
		cs.Lines = nil
		cs.IsLineDuring = false
		cs.Forced = true
	})

	done := make(chan struct{}, 1)
	exCandlesCh := make(chan cdl.Candle, 1)

	go func() {
		for {
			select {
			case candle := <-exCandlesCh:

				timeout := time.After(
					time.Second * time.Duration(t.Interval.AsSeconds()-10),
				)

			loop:
				for {
					select {
					case <-timeout:
						break loop
					case <-done:
						break loop
					default:
						ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
						exCandles, err := b.broker.ExtendStartCandles(
							ctx,
							[]cdl.Candle{candle},
							t.Category,
							t.Symbol,
							t.Interval,
							t.Limit,
						)

						cancel()

						if err == nil {
							b.Push(func(s *State) {
								cs := s.Chart[t.Idx]
								n := len(cs.Candles)
								if n > 0 {
									last := cs.Candles[n-1]
									cs.Candles = exCandles
									cs.Candles = append(cs.Candles, last)
								}
								cs.Forced = true
							})
							break loop
						}
					}
				}

			case <-done:
				return

			}
		}
	}()

	exCandlesCh <- first.Candle

	go func() {
		defer close(done)

		intervalMs := int64(t.Interval.AsMilli() + 999)

		for d := range sub.C {
			if d.Confirm {
				b.Push(func(s *State) {
					cs := s.Chart[t.Idx]

					n := len(cs.Candles) - 1
					if n > 1 {
						if cs.Candles[n-1].Time-cs.Candles[n-2].Time > intervalMs {
							exCandlesCh <- d.Candle
						}
					}

					cs.Candles = append(cs.Candles, d.Candle)
					cs.Forced = true
				})
				continue
			}

			b.Push(func(s *State) {
				cs := s.Chart[t.Idx]

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
		obS.Category = t.Category.AsString(true)
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

type Background struct {
	Tx chan Task

	commit atomic.Pointer[CommitFn]

	broker broker.Broker

	stream        [MAX_CATEGORY_STREAMS]broker.Stream // public streams
	privateStream broker.PrivateStream

	chartSub     [MAX_TOPIC_SUBSCRIPTIONS]*broker.Subscription[cdl.CandleStreamData]
	orderBookSub [MAX_TOPIC_SUBSCRIPTIONS]*broker.Subscription[[2][][2]float64]
}

func InitBackground(ctx context.Context, b broker.Broker) *Background {
	bg := &Background{
		Tx:     make(chan Task, 1),
		broker: b,
	}

	go func() {
		for {
			select {
			case t := <-bg.Tx:
				if err := t.Execute(bg); err != nil {
					bg.Push(
						CommitCommandLineError(err.Error()),
					)
				}
			case <-ctx.Done():
				bg.privateStream.Close()
				for _, s := range bg.stream {
					if s == nil {
						continue
					}
					s.Close()
				}
				return
			}
		}
	}()

	return bg
}

func (b *Background) GetOrInitStream(c broker.Category) broker.Stream {

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

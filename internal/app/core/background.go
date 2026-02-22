package core

import (
	"context"
	"nlkli/raytrade/internal/broker"
	"nlkli/raytrade/internal/cdl"
	"nlkli/raytrade/internal/ws"
	"sync"
	"sync/atomic"
	"time"
)

type Task any

// Commands

// sub <chart> <i> <F.BTCUSDT.15>
// sub <orderbook> <i> <F.BTCUSDT.depth>

// unsub <chart> <i>
// unsub <orderbook> <i>

type SubChart struct {
	Idx      int
	Category broker.Category
	Symbol   string
	Interval cdl.Interval
	Limit    int
}

func (t *SubChart) Execute(b *Background) error {

	stream := b.CreateStream(t.Category, nil)

	sub, err := stream.SubscribeCandle(t.Symbol, t.Interval)
	if err != nil {
		return err
	}

	first := <-sub.C

	b.Push(func(s *State) {
		cs := s.Chart[t.Idx]
		cs.SecInterval = float32(t.Interval.AsSeconds())
		cs.Candles = []cdl.Candle{first.Candle}
	})

	if b.chartSub[t.Idx] != nil {
		b.chartSub[t.Idx].Stop()
	}
	b.chartSub[t.Idx] = sub

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
						ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
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
								cs.Candles = append(
									exCandles,
									cs.Candles[min(1, len(cs.Candles)):]...,
								)

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

		intervalMs := int64(t.Interval.AsMilli() + 2_000)

		for d := range sub.C {
			if d.Confirm {
				b.Push(func(s *State) {
					cs := s.Chart[t.Idx]

					if d.Candle.Time-cs.Candles[len(cs.Candles)-1].Time > intervalMs { // TODO?
						exCandlesCh <- d.Candle
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

	stream := b.CreateStream(t.Category, nil)

	sub, err := stream.SubscribeOrderBook(t.Symbol, t.Depth)
	if err != nil {
		return err
	}

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

	stream        []broker.Stream
	privateStream broker.PrivateStream

	chartSub     []*broker.Subscription[cdl.CandleStreamData]
	orderBookSub []*broker.Subscription[[2][][2]float64]

	mu sync.Mutex
}

func (b *Background) CreateStream(

	c broker.Category,
	onConnected ws.OnConnectedFn,
	opts ...ws.PolicyOption,

) broker.Stream {

	stream := b.stream[c]

	if stream == nil {
		stream = b.broker.CreateStream(c, onConnected, opts...)
		b.stream[c] = stream
		return stream
	}

	return stream
}

func InitBackground(ctx context.Context, br broker.Broker) *Background {
	b := &Background{
		Tx:           make(chan Task, 32),
		broker:       br,
		stream:       make([]broker.Stream, 16),
		chartSub:     make([]*broker.Subscription[cdl.CandleStreamData], 16),
		orderBookSub: make([]*broker.Subscription[[2][][2]float64], 16),
	}

	go func() {
		for {
			select {
			case t := <-b.Tx:
				switch t := t.(type) {
				case SubChart:
					t.Execute(b)
				case SubOrderBook:
					t.Execute(b)
				default:
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	return b
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

package core

import (
	"context"
	"nlkli/raytrade/internal/broker"
	"nlkli/raytrade/internal/cdl"
	"sync"
	"sync/atomic"
)

type Task any

type InstrumentObserverT struct {
	Category         broker.Category
	Symbol           string
	Interval         cdl.Interval
	InitCandlesLimit int
}

func (t *InstrumentObserverT) run(b *Background) {
	if b.stream == nil {
		b.stream = b.broker.CreateStream(t.Category, nil)
	}

	chartSub, err := b.stream.SubscribeCandle(t.Symbol, t.Interval)
	if err != nil {
		b.Push(CommitCommandLineErrorAnd(
			err.Error(),
			func(s *State) {
				s.StatusLine.Symbol = s.Bg.Symbol
				s.StatusLine.Interval = s.Bg.Interval.AsString()
			},
		))
		return
	}

	var first cdl.CandleStreamData
	for {
		first = <-chartSub.C
		if !first.Confirm {
			break
		}
	}

	candles := []cdl.Candle{first.Candle}

	candles, err = b.broker.ExtendStartCandles(
		context.TODO(),
		candles,
		t.Category,
		t.Symbol,
		t.Interval,
		t.InitCandlesLimit,
	)
	if err != nil {
		b.Push(CommitCommandLineErrorAnd(
			err.Error(),
			func(s *State) {
				s.StatusLine.Symbol = s.Bg.Symbol
				s.StatusLine.Interval = s.Bg.Interval.AsString()
			},
		))
		chartSub.Stop()
		return
	}

	if b.chartSub != nil {
		b.chartSub.Stop()
	}
	b.wg.Wait()
	b.chartSub = chartSub

	var f CommitFn
	f = func(s *State) {
		s.Bg.IsActiveIO = true
		s.Bg.Category = t.Category
		s.Bg.Symbol = t.Symbol
		s.Bg.Interval = t.Interval

		s.StatusLine.Symbol = t.Symbol
		s.StatusLine.Interval = t.Interval.AsString()

		s.Chart[0].SecInterval = float32(t.Interval.AsSeconds())
		s.Chart[0].Candles = candles
		s.Chart[0].Forced = true
	}

	b.Push(f)

	obSub, _ := b.stream.SubscribeOrderBook(t.Symbol, 200)

	b.wg.Go(func() {
		for ob := range obSub.C {
			b.Push(
				func(s *State) {
					s.OrderBook[0].Bids = ob[0]
					s.OrderBook[0].Asks = ob[1]
					s.OrderBook[0].Forced = true
				},
			)
		}
	})

	b.wg.Go(func() {
		defer b.Push(func(s *State) {
			s.Bg.IsActiveIO = false
		})

		for d := range chartSub.C {
			var f CommitFn
			if d.Confirm {
				f = func(s *State) {
					s.Chart[0].Candles = append(s.Chart[0].Candles, d.Candle)
					s.Chart[0].Forced = true
				}
			} else {
				f = func(s *State) {
					s.Chart[0].Candles[len(s.Chart[0].Candles)-1] = d.Candle
					s.Chart[0].MaxP = max(s.Chart[0].MaxP, d.Candle.H)
					s.Chart[0].MinP = min(s.Chart[0].MinP, d.Candle.L)
					s.Chart[0].MidP = (s.Chart[0].MaxP + s.Chart[0].MinP) * .5
					s.Chart[0].RngP = s.Chart[0].MaxP - s.Chart[0].MinP
				}
			}
			b.Push(f)
		}
	})
}

type Background struct {
	Tx chan Task

	commit atomic.Pointer[CommitFn]

	broker   broker.Broker
	stream   broker.BrokerStream
	chartSub *broker.BrokerStreamSubscription[cdl.CandleStreamData]

	mu sync.Mutex
	wg sync.WaitGroup

	doneIOT chan struct{}
}

func InitBackground(ctx context.Context, br broker.Broker) *Background {
	b := &Background{
		Tx:     make(chan Task, 32),
		broker: br,
	}

	go func() {
		for {
			select {
			case t := <-b.Tx:
				switch t := t.(type) {
				case InstrumentObserverT:
					t.run(b)
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

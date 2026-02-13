package app

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
	doneIOT := make(chan struct{}, 1)

	stream, err := b.broker.CandleStream(doneIOT, t.Category, t.Symbol, t.Interval)
	if err != nil {
		b.push(CommitCommandLineErrorAnd(
			err.Error(),
			func(s *State) {
				s.StatusLine.Symbol = s.Bg.Symbol
				s.StatusLine.Interval = s.Bg.Interval.AsString()
			},
		))
		close(doneIOT)
		return
	}

	var first cdl.CandleStreamData
	for {
		first = <-stream
		if !first.Confirm {
			break
		}
	}

	candles := []cdl.Candle{first.Candle}

	candles, err = b.broker.ExtendStartCandles(
		candles,
		t.Category,
		t.Symbol,
		t.Interval,
		t.InitCandlesLimit,
	)
	if err != nil {
		b.push(CommitCommandLineErrorAnd(
			err.Error(),
			func(s *State) {
				s.StatusLine.Symbol = s.Bg.Symbol
				s.StatusLine.Interval = s.Bg.Interval.AsString()
			},
		))
		doneIOT <- struct{}{}
		return
	}

	if b.doneIOT != nil {
		close(b.doneIOT)
	}
	b.wg.Wait()
	b.doneIOT = doneIOT

	var f CommitFn
	f = func(s *State) {
		s.Bg.IsActiveIO = true
		s.Bg.Category = t.Category
		s.Bg.Symbol = t.Symbol
		s.Bg.Interval = t.Interval

		s.StatusLine.Symbol = t.Symbol
		s.StatusLine.Interval = t.Interval.AsString()

		s.Chart.SecInterval = float32(t.Interval.AsSeconds())
		s.Chart.Candles = candles
		s.Chart.Forced = true
	}

	b.push(f)

	b.wg.Go(func() {
		defer b.push(func(s *State) {
			s.Bg.IsActiveIO = false
		})

		for d := range stream {
			var f CommitFn
			if d.Confirm {
				f = func(s *State) {
					s.Chart.Candles = append(s.Chart.Candles, d.Candle)
					s.Chart.Forced = true
				}
			} else {
				f = func(s *State) {
					s.Chart.Candles[len(s.Chart.Candles)-1] = d.Candle
					s.Chart.MaxP = max(s.Chart.MaxP, d.Candle.H)
					s.Chart.MinP = min(s.Chart.MinP, d.Candle.L)
					s.Chart.MidP = (s.Chart.MaxP + s.Chart.MinP) * .5
					s.Chart.RngP = s.Chart.MaxP - s.Chart.MinP
				}
			}
			b.push(f)
		}
	})
}

type Background struct {
	Tx chan Task

	buff [4]CommitFn
	head atomic.Uint32
	tail atomic.Uint32

	broker broker.Broker

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

func (b *Background) Update(s *State) {
	tail := b.tail.Load()
	head := b.head.Load()

	for tail != head {
		b.buff[tail&3](s)
		b.buff[tail&3] = nil
		tail++
	}

	b.tail.Store(tail)
}

func (b *Background) push(f CommitFn) bool {
	head := b.head.Load()
	tail := b.tail.Load()

	if head-tail == 4 {
		return false
	}

	b.buff[head&3] = f
	b.head.Store(head + 1)

	return true
}

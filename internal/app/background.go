package app

import (
	"context"
	"fmt"
	"nlkli/raytrade/internal/broker"
	"nlkli/raytrade/internal/cdl"
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
	done := make(chan struct{}, 1)
	stream, err := b.broker.CandleStream(done, t.Category, t.Symbol, t.Interval)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	firs := <-stream
	candles := []cdl.Candle{firs.Candle}

	candles, err = b.broker.ExtendStartCandles(
		candles,
		t.Category,
		t.Symbol,
		t.Interval,
		t.InitCandlesLimit,
	)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	var f CommitFn
	f = func(s *State) error {
		s.Chart.Candles = candles
		s.Chart.Forced = true
		return nil
	}

	b.push(f)

	go func() {
		for d := range stream {
			var f CommitFn
			if d.Confirm {
				f = func(s *State) error {
					s.Chart.Candles = append(s.Chart.Candles, d.Candle)
					s.Chart.Forced = true
					return nil
				}
			} else {
				f = func(s *State) error {
					s.Chart.Candles[len(s.Chart.Candles)-1] = d.Candle
					return nil
				}
			}
			b.push(f)
		}
	}()
}

type Background struct {
	Tx chan Task

	buf  [4]CommitFn
	head atomic.Uint32
	tail atomic.Uint32

	broker broker.Broker
}

func InitBackground(ctx context.Context, br broker.Broker) *Background {
	b := &Background{
		Tx: make(chan Task, 32),

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

func (b *Background) Update(s *State) error {
	tail := b.tail.Load()
	head := b.head.Load()

	for tail != head {
		f := b.buf[tail&3]
		b.buf[tail&3] = nil

		if err := f(s); err != nil {
			b.tail.Store(tail + 1)
			return err
		}

		tail++
	}

	b.tail.Store(tail)
	return nil
}

func (b *Background) push(f CommitFn) bool {
	head := b.head.Load()
	tail := b.tail.Load()

	if head-tail == 4 {
		return false // full
	}

	b.buf[head&3] = f
	b.head.Store(head + 1)

	return true
}

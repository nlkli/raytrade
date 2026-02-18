package broker

import (
	"context"
	"nlkli/raytrade/internal/cdl"
	"nlkli/raytrade/internal/ws"
	"sync"
)

type Category int

const (
	Spot Category = iota
	Futures
)

type Broker interface {
	GetCandles(
		ctx context.Context,
		category Category,
		symbol string,
		interval cdl.Interval,
		limit int,
		start *int,
		end *int,
	) ([]cdl.Candle, error)

	// Create a new candles slice
	ExtendStartCandles(
		ctx context.Context,
		candles []cdl.Candle,
		category Category,
		symbol string,
		interval cdl.Interval,
		limit int,
	) ([]cdl.Candle, error)

	// Create a new candles slice
	ExtendEndCandles(
		ctx context.Context,
		candles []cdl.Candle,
		category Category,
		symbol string,
		interval cdl.Interval,
		limit int,
	) ([]cdl.Candle, error)

	// [bids, asks][][price, size]
	GetOrderBook(
		ctx context.Context,
		category Category,
		symbol string,
		limit int,
	) (*[2][][2]float64, error)

	CreateStream(
		category Category,
		onConnected ws.OnConnectedFn,
		opts ...ws.PolicyOption,
	) BrokerStream
}

type BrokerStreamSubscription[T any] struct {
	C      <-chan T
	onStop func() error
	once   sync.Once
	err    error
}

func NewBrokerStreamSubscription[T any](
	ch <-chan T, onStop func() error,
) *BrokerStreamSubscription[T] {
	return &BrokerStreamSubscription[T]{
		C:      ch,
		onStop: onStop,
	}
}

func (s *BrokerStreamSubscription[T]) Stop() error {
	s.once.Do(func() {
		if s.onStop != nil {
			s.err = s.onStop()
		}
	})

	return s.err
}

type BrokerStream interface {
	SubscribeCandle(
		symbol string,
		interval cdl.Interval,
	) (*BrokerStreamSubscription[cdl.CandleStreamData], error)

	SubscribeOrderBook(
		symbol string, depth int,
	) (*BrokerStreamSubscription[[2][][2]float64], error)
}

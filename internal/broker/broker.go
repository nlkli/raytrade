package broker

import (
	"context"
	"nlkli/raytrade/internal/cdl"
	"nlkli/raytrade/internal/ws"
	"sync"
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

	GetPosition(
		ctx context.Context,
		category Category,
	) ([]Position, error)

	GetOpenOrders(
		ctx context.Context,
		category Category,
		symbol string,
		limit int,
	) ([]Order, string, error)

	CreateStream(
		category Category,
		onConnected ws.OnConnectedFn,
		opts ...ws.PolicyOption,
	) Stream

	CreatePrivateStream(
		onConnected ws.OnConnectedFn,
		opts ...ws.PolicyOption,
	) PrivateStream
}

func SplitSubscription[T any](

	sub *Subscription[T],
	onStop func() error,

) (*Subscription[T], *Subscription[T]) {
	subCh1, subCh2 := make(chan T, 1), make(chan T, 1)
	sub1, sub2 := NewBrokerStreamSubscription(subCh1, sub.onStop),
		NewBrokerStreamSubscription(subCh2, onStop)

	go func() {
		defer close(subCh1)
		defer close(subCh2)

		for d := range sub.C {
			subCh1 <- d
			subCh2 <- d
		}
	}()

	return sub1, sub2
}

type Subscription[T any] struct {
	C      <-chan T
	onStop func() error
	once   sync.Once
	err    error
}

func NewBrokerStreamSubscription[T any](
	ch <-chan T, onStop func() error,
) *Subscription[T] {
	return &Subscription[T]{
		C:      ch,
		onStop: onStop,
	}
}

func (s *Subscription[T]) Stop() error {
	s.once.Do(func() {
		if s.onStop != nil {
			s.err = s.onStop()
		}
	})

	return s.err
}

type Stream interface {
	SubscribeCandle(
		symbol string,
		interval cdl.Interval,
	) (*Subscription[cdl.CandleStreamData], error)

	SubscribeOrderBook(
		symbol string, depth int,
	) (*Subscription[[2][][2]float64], error)

	Close()
}

type PrivateStream interface {
	SubscribePosition() (*Subscription[Position], error)
	SubscribeOrder() (*Subscription[Order], error)

	Close()
}

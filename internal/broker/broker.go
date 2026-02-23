package broker

import (
	"context"
	"errors"
	"nlkli/raytrade/internal/cdl"
	"nlkli/raytrade/internal/ws"
	"strings"
	"sync"
	"time"
)

type OrderBookDT [2][][2]float64

type Category int

const (
	Spot Category = iota
	Futures
)

type Side string

const (
	Long  Side = "Long"
	Short Side = "Short"
)

type OrderType string

const (
	Market OrderType = "Market"
	Limit  OrderType = "Limit"
)

type OrderStatus string

const (
	Open   OrderStatus = "Open"
	Closed OrderStatus = "Closed"
)

type Order struct {
	Symbol     string
	Side       Side
	Status     OrderStatus
	Price      float64
	Qty        float64
	ExecQty    float64
	ExecValue  float64
	EntryPrice *float64
	CreatedAt  time.Time
}

type Position struct {
	Symbol     string
	Side       Side
	Size       float64
	EntryPrice float64
	CreatedAt  time.Time
}

func CategoryFromString(s string) (Category, error) {
	s = strings.ToUpper(s)

	switch s {
	case "S", "SP", "SPOT":
		return Spot, nil
	case "F", "FT", "FUTURES":
		return Futures, nil
	}

	return -1, errors.New("invalid category string")
}

func CategoryToString(c Category, short bool) string {
	switch c {
	case Spot:
		if short {
			return "S"
		}
		return "Spot"
	case Futures:
		if short {
			return "F"
		}
		return "Futures"
	default:
		return ""
	}
}

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
		symbol string,
	) (*Position, error)

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
}

type PrivateStream interface {
	SubscribePosition() (*Subscription[Position], error)
}

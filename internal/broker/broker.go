package broker

import (
	"context"
	"errors"
	"nlkli/raytrade/internal/cdl"
	"nlkli/raytrade/internal/ws"
	"strings"
	"sync"
)

type OrderBookDT [2][][2]float64

type Category int

const (
	Spot Category = iota
	Futures
)

type Position struct {
	// Идентификация
	Symbol      string `json:"symbol"`      // Тикер
	Side        string `json:"side"`        // "Buy" или "Sell"
	PositionIdx int    `json:"positionIdx"` // 0: One-Way, 1: Buy side, 2: Sell side

	// Размер и цена
	Size       float64 `json:"size"`       // Размер позиции (положительное число)
	EntryPrice float64 `json:"entryPrice"` // Средняя цена входа
	MarkPrice  float64 `json:"markPrice"`  // Текущая рыночная цена
	Leverage   float64 `json:"leverage"`   // Кредитное плечо

	// P&L
	UnrealisedPnl  float64 `json:"unrealisedPnl"`  // Нереализованная прибыль/убыток
	CumRealisedPnl float64 `json:"cumRealisedPnl"` // Накопленная реализованная прибыль

	// Риски
	LiqPrice   float64 `json:"liqPrice"`   // Цена ликвидации
	PositionIM float64 `json:"positionIM"` // Начальная маржа
	PositionMM float64 `json:"positionMM"` // Поддерживающая маржа

	// Стоп-ордера
	TakeProfit float64 `json:"takeProfit"` // Цена тейк-профита
	StopLoss   float64 `json:"stopLoss"`   // Цена стоп-лосса

	// Статус
	PositionStatus string `json:"positionStatus"` // "Normal", "Liq", "Adl"
	IsReduceOnly   bool   `json:"isReduceOnly"`   // Только уменьшение позиции
	AutoAddMargin  bool   `json:"autoAddMargin"`  // Автодобавление маржи

	// Временные метки
	CreatedAt int64 `json:"createdAt"` // Время создания позиции
	UpdatedAt int64 `json:"updatedAt"` // Время последнего обновления

	// Категория продукта (для различения типов)
	Category string `json:"category"` // "linear", "inverse", "option"
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
}

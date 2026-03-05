package broker

import (
	"errors"
	"strings"
	"time"
)

type OrderBookDT [2][][2]float64

type Category int

const (
	Spot Category = iota
	Futures
)

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

func (c Category) AsString(short bool) string {
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
	Category   Category
	Symbol     string
	Side       Side
	Status     OrderStatus
	Price      float64
	Qty        float64
	ExecQty    float64
	ExecValue  float64
	EntryPrice float64
	CreatedAt  time.Time
}

type Position struct {
	Category       Category
	Symbol         string
	Side           Side
	Size           float64
	EntryPrice     float64
	PositionValue  float64
	PositionIM     float64
	Leverage       int
	MarkPrice      float64
	BreakEvenPrice float64
	UnrealisedPnl  float64
	RealisedPnl    float64
	LiqPrice       float64
	TakeProfit     float64
	StopLoss       float64
	CreatedAt      time.Time
}

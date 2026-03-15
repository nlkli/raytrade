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

type Side int

const (
	Long Side = iota
	Short
)

type OrderType int

const (
	Market OrderType = iota
	Limit
)

type OrderStatus int

const (
	Open OrderStatus = iota
	Closed
)

type MarketUnit int

const (
	BaseCoin MarketUnit = iota
	QuoteCoin
)

type TriggerBy int

const (
	LastPrice TriggerBy = iota
	IndexPrice
	MarkPrice
)

type TriggerDirection int

const (
	Rise TriggerDirection = 1
	Fall TriggerDirection = 2
)

type TpslMode int

const (
	Full    TpslMode = iota
	Partial TpslMode = iota
)

// https://bybit-exchange.github.io/docs/v5/order/create-order
type PlaceOrderParams struct {
	Category         Category
	Symbol           string
	IsLeverage       *bool
	Side             Side
	Type             OrderType
	Qty              float64
	MarketUnit       *MarketUnit
	Price            *float64
	TriggerDirection *TriggerDirection
	TriggerPrice     *float64
	TriggerBy        *TriggerBy
	TakeProfit       *float64
	StopLoss         *float64
	TpTriggerBy      *TriggerBy
	SlTriggerBy      *TriggerBy
	ReduceOnly       *bool
	CloseOnTrigger   *bool
	TpslMode         *TpslMode
	TpLimitPrice     *float64
	SlLimitPrice     *float64
	TpOrderType      *OrderType
	SlOrderType      *OrderType
}

// https://bybit-exchange.github.io/docs/v5/order/open-order
type Order struct {
	Category         Category
	Symbol           string
	Id               string
	LinkId           string
	Price            float64
	Qty              float64
	MarketUnit       MarketUnit
	Side             Side
	IsLeverage       bool
	Status           OrderStatus
	Type             OrderType
	AvgPrice         float64
	LeavesQty        float64
	LeavesValue      float64
	ExecQty          float64
	ExecValue        float64
	TriggerPrice     float64
	TakeProfit       float64
	StopLoss         float64
	TpslMode         TpslMode
	TpLimitPrice     float64
	SlLimitPrice     float64
	TpTriggerBy      TriggerBy
	SlTriggerBy      TriggerBy
	TriggerDirection TriggerDirection
	TriggerBy        TriggerBy
	ReduceOnly       bool
	CloseOnTrigger   bool
	CreatedAt        time.Time
}

// https://bybit-exchange.github.io/docs/v5/position
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

type Execution struct {
	Category    Category
	Symbol      string
	Side        Side
	Qty         float64
	Price       float64
	OrderId     string
	OrderLinkId string
	IsMaker     bool
	Time        time.Time
}

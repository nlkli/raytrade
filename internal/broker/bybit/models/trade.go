package models

import "encoding/json"

type OrderResult struct {
	OrderId     string `json:"orderId"`
	OrderLinkId string `json:"orderLinkId"` // User customised order ID
}

type CancelAllOrdersResult struct {
	List    []OrderResult
	Success string
}

type OrderListResult struct {
	Category       string      `json:"category"`
	NextPageCursor string      `json:"nextPageCursor"`
	List           []OrderInfo `json:"list"`
}

type OrderInfo struct {
	OrderId            string           `json:"orderId"`
	OrderLinkId        string           `json:"orderLinkId"`
	ParentOrderLinkId  string           `json:"parentOrderLinkId"`
	BlockTradeId       string           `json:"blockTradeId"`
	Symbol             string           `json:"symbol"`
	Price              string           `json:"price"`
	Qty                string           `json:"qty"`
	Side               Side             `json:"side"`
	IsLeverage         string           `json:"isLeverage"`
	PositionIdx        PositionIdx      `json:"positionIdx"`
	OrderStatus        OrderStatus      `json:"orderStatus"`
	CreateType         string           `json:"createType"`
	CancelType         string           `json:"cancelType"`
	RejectReason       string           `json:"rejectReason"`
	AvgPrice           string           `json:"avgPrice"`
	LeavesQty          string           `json:"leavesQty"`
	LeavesValue        string           `json:"leavesValue"`
	CumExecQty         string           `json:"cumExecQty"`
	CumExecValue       string           `json:"cumExecValue"`
	CumExecFee         string           `json:"cumExecFee"`
	TimeInForce        TimeInForce      `json:"timeInForce"`
	OrderType          OrderType        `json:"orderType"`
	StopOrderType      string           `json:"stopOrderType"`
	OrderIv            string           `json:"orderIv"`
	MarketUnit         MarketUnit       `json:"marketUnit"`
	TriggerPrice       string           `json:"triggerPrice"`
	TakeProfit         string           `json:"takeProfit"`
	StopLoss           string           `json:"stopLoss"`
	TpslMode           TpslMode         `json:"tpslMode"`
	OcoTriggerBy       string           `json:"ocoTriggerBy"`
	TpLimitPrice       string           `json:"tpLimitPrice"`
	SlLimitPrice       string           `json:"slLimitPrice"`
	TpTriggerBy        TpSlTriggerBy    `json:"tpTriggerBy"`
	SlTriggerBy        TpSlTriggerBy    `json:"slTriggerBy"`
	TriggerDirection   TriggerDirection `json:"triggerDirection"`
	TriggerBy          TriggerBy        `json:"triggerBy"`
	LastPriceOnCreated string           `json:"lastPriceOnCreated"`
	BasePrice          string           `json:"basePrice"`
	ReduceOnly         bool             `json:"reduceOnly"`
	CloseOnTrigger     bool             `json:"closeOnTrigger"`
	PlaceType          string           `json:"placeType"`
	SmpType            SMPType          `json:"smpType"`
	SmpGroup           int              `json:"smpGroup"`
	SmpOrderId         string           `json:"smpOrderId"`
	CreatedTime        string           `json:"createdTime"`
	UpdatedTime        string           `json:"updatedTime"`
	CumFeeDetail       json.RawMessage  `json:"cumFeeDetail"`
}

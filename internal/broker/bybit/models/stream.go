package models

import "encoding/json"

type StreamOp = string

const (
	StreamOpAuth        StreamOp = "auth"
	StreamOpSubscribe   StreamOp = "subscribe"
	StreamOpUnsubscribe StreamOp = "unsubscribe"
)

type StreamOpRequest struct {
	ReqID string `json:"req_id,omitempty"`
	Op    string `json:"op"`
	Args  []any  `json:"args"`
}

// https://bybit-exchange.github.io/docs/v5/ws/connect
// StreamOpResult represents an operation result (auth, subscribe, unsubscribe).
type StreamOpResult struct {
	Success bool   `json:"success"`
	RetMsg  string `json:"ret_msg"`
	ConnID  string `json:"conn_id"`
	ReqID   string `json:"req_id,omitempty"`
	Op      string `json:"op"`
}

// StreamData wraps a topic-specific stream payload.
type StreamData struct {
	Topic string          `json:"topic"`
	Type  string          `json:"type,omitempty"`
	TS    int64           `json:"ts"`
	Data  json.RawMessage `json:"data"`
}

// https://bybit-exchange.github.io/docs/v5/websocket/public/kline
// StreamKlineFrame represents a single candlestick.
type StreamKlineFrame struct {
	Start     int64
	End       int64
	Interval  string
	Open      string
	Close     string
	High      string
	Low       string
	Volume    string
	Turnover  string
	Confirm   bool
	Timestamp int64
}

// StreamKlineData is a batch of KlineFrame from the stream.
type StreamKlineData = []StreamKlineFrame

// https://bybit-exchange.github.io/docs/v5/websocket/public/orderbook
type StreamOrderBookData struct {
	Symbol   string      `json:"s"`
	Bids     [][2]string `json:"b"`   // Desc sorted
	Asks     [][2]string `json:"a"`   // Asc sorted
	UpdateID int64       `json:"u"`   // Update ID
	Seq      int64       `json:"seq"` // Cross sequence
	CTS      int64       `json:"cts"` // Matching engine timestamp
}

type StreamPositionInfo struct {
	Category               string `json:"category"`
	Symbol                 string `json:"symbol"`
	Side                   string `json:"side"`
	Size                   string `json:"size"`
	PositionIdx            int    `json:"positionIdx"`
	PositionValue          string `json:"positionValue"`
	RiskID                 int    `json:"riskId"`
	RiskLimitValue         string `json:"riskLimitValue"`
	EntryPrice             string `json:"entryPrice"`
	MarkPrice              string `json:"markPrice"`
	Leverage               string `json:"leverage"`
	BreakEvenPrice         string `json:"breakEvenPrice"`
	AutoAddMargin          int    `json:"autoAddMargin"`
	PositionIM             string `json:"positionIM"`
	PositionMM             string `json:"positionMM"`
	LiqPrice               string `json:"liqPrice"`
	TakeProfit             string `json:"takeProfit"`
	StopLoss               string `json:"stopLoss"`
	TrailingStop           string `json:"trailingStop"`
	UnrealisedPnl          string `json:"unrealisedPnl"`
	CurRealisedPnl         string `json:"curRealisedPnl"`
	SessionAvgPrice        string `json:"sessionAvgPrice"`
	Delta                  string `json:"delta"`
	Gamma                  string `json:"gamma"`
	Vega                   string `json:"vega"`
	Theta                  string `json:"theta"`
	CumRealisedPnl         string `json:"cumRealisedPnl"`
	PositionStatus         string `json:"positionStatus"`
	AdlRankIndicator       int    `json:"adlRankIndicator"`
	IsReduceOnly           bool   `json:"isReduceOnly"`
	CreatedTime            string `json:"createdTime"`
	UpdatedTime            string `json:"updatedTime"`
	Seq                    string `json:"seq"`
	MmrSysUpdatedTime      string `json:"mmrSysUpdatedTime"`
	LeverageSysUpdatedTime string `json:"leverageSysUpdatedTime"`
	PositionIMByMp         string `json:"positionIMByMp"`
	PositionMMByMp         string `json:"positionMMByMp"`
	TpslMode               string `json:"tpslMode"`
	BustPrice              string `json:"bustPrice"`
	PositionBalance        string `json:"positionBalance"`
	TradeMode              int    `json:"tradeMode"` // Deprecated
}

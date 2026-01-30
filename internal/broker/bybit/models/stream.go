package models

import "encoding/json"

type StreamOperation = string

const (
	StreamOpAuth        StreamOperation = "auth"
	StreamOpSubscribe   StreamOperation = "subscribe"
	StreamOpUnsubscribe StreamOperation = "unsubscribe"
)

type StreamOperationRequest struct {
	ReqID string `json:"req_id,omitempty"`
	Op    string `json:"op"`
	Args  []any  `json:"args"`
}

// StreamEnvelopeMessage is a lightweight discriminator for incoming WS messages.
// It is used to detect message type before full decoding.
type StreamEnvelopeMessage struct {
	Topic   string `json:"topic,omitempty"`
	Success *bool  `json:"success,omitempty"`
}

// https://bybit-exchange.github.io/docs/v5/ws/connect
// StreamOperationResult represents an operation result (auth, subscribe, unsubscribe).
type StreamOperationResult struct {
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

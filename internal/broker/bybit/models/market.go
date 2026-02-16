package models

type KlineResult struct {
	Category string      `json:"category"`
	Symbol   string      `json:"symbol"`
	List     [][7]string `json:"list"` // [[time, open, high, low, close, vol, turnover], ...]
}

type OrderBookResult struct {
	S   string      `json:"s"`   // Symbol
	B   [][2]string `json:"b"`   // Bids: [[price, size], ...]
	A   [][2]string `json:"a"`   // Asks: [[price, size], ...]
	Ts  int64       `json:"ts"`  // System timestamp (ms)
	U   int64       `json:"u"`   // Update ID
	Seq int64       `json:"seq"` // Cross sequence
	Cts int64       `json:"cts"` // Matching engine timestamp (ms)
}

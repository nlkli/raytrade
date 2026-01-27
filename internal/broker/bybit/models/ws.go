package models

import "encoding/json"

type OperationResponse struct {
	Success bool   `json:"success"`
	RetMsg  string `json:"ret_msg"`
	ConnID  string `json:"conn_id"`
	ReqID   string `json:"req_id,omitempty"`
	Op      string `json:"op"`
}

type StreamDataMessage struct {
	Topic string          `json:"topic"`
	Type  string          `json:"type,omitempty"`
	TS    int64           `json:"ts"`
	Data  json.RawMessage `json:"data"`
}

type KlineFrame struct {
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

type KlineStreamData = []KlineFrame

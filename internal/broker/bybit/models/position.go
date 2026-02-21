package models

type PositionInfoResult struct {
	Category       string         `json:"category"`
	NextPageCursor string         `json:"nextPageCursor"`
	List           []PositionInfo `json:"list"`
}

type PositionInfo struct {
	PositionIdx            PositionIdx            `json:"positionIdx"`
	RiskID                 int            `json:"riskId"`
	RiskLimitValue         string         `json:"riskLimitValue"`
	Symbol                 string         `json:"symbol"`
	Side                   string         `json:"side"`
	Size                   string         `json:"size"`
	AvgPrice               string         `json:"avgPrice"`
	PositionValue          string         `json:"positionValue"`
	AutoAddMargin          int            `json:"autoAddMargin"`
	PositionStatus         PositionStatus `json:"positionStatus"`
	Leverage               string         `json:"leverage"`
	BreakEvenPrice         string         `json:"breakEvenPrice"`
	MarkPrice              string         `json:"markPrice"`
	LiqPrice               string         `json:"liqPrice"`
	PositionIM             string         `json:"positionIM"`
	PositionIMByMp         string         `json:"positionIMByMp"`
	PositionMM             string         `json:"positionMM"`
	PositionMMByMp         string         `json:"positionMMByMp"`
	TakeProfit             string         `json:"takeProfit"`
	StopLoss               string         `json:"stopLoss"`
	TrailingStop           string         `json:"trailingStop"`
	SessionAvgPrice        string         `json:"sessionAvgPrice"`
	Delta                  string         `json:"delta"`
	Gamma                  string         `json:"gamma"`
	Vega                   string         `json:"vega"`
	Theta                  string         `json:"theta"`
	UnrealisedPnl          string         `json:"unrealisedPnl"`
	CurRealisedPnl         string         `json:"curRealisedPnl"`
	CumRealisedPnl         string         `json:"cumRealisedPnl"`
	AdlRankIndicator       int            `json:"adlRankIndicator"`
	CreatedTime            string         `json:"createdTime"`
	UpdatedTime            string         `json:"updatedTime"`
	Seq                    string         `json:"seq"`
	IsReduceOnly           bool           `json:"isReduceOnly"`
	MmrSysUpdatedTime      string         `json:"mmrSysUpdatedTime"`
	LeverageSysUpdatedTime string         `json:"leverageSysUpdatedTime"`
	TpslMode               string         `json:"tpslMode"`
	BustPrice              string         `json:"bustPrice"`
	PositionBalance        string         `json:"positionBalance"`
	TradeMode              int            `json:"tradeMode"` // Deprecated
}

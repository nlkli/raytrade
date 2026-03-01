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

type InstrumentInfoResult struct {
}

type InstrumentInfoSpot struct {
	Symbol        string `json:"symbol"`
	BaseCoin      string `json:"baseCoin"`
	QuoteCoin     string `json:"quoteCoin"`
	Innovation    string `json:"innovation"`
	Status        string `json:"status"`
	MarginTrading string `json:"marginTrading"`
	StTag         string `json:"stTag"`
	LotSizeFilter struct {
		BasePrecision             string `json:"basePrecision"`
		QuotePrecision            string `json:"quotePrecision"`
		MinOrderQty               string `json:"minOrderQty"`
		MaxOrderQty               string `json:"maxOrderQty"`
		MinOrderAmt               string `json:"minOrderAmt"`
		MaxOrderAmt               string `json:"maxOrderAmt"`
		MaxLimitOrderQty          string `json:"maxLimitOrderQty"`
		MaxMarketOrderQty         string `json:"maxMarketOrderQty"`
		PostOnlyMaxLimitOrderSize string `json:"postOnlyMaxLimitOrderSize"`
	} `json:"lotSizeFilter"`
	PriceFilter struct {
		TickSize string `json:"tickSize"`
	} `json:"priceFilter"`
	RiskParameters struct {
		PriceLimitRatioX string `json:"priceLimitRatioX"`
		PriceLimitRatioY string `json:"priceLimitRatioY"`
	} `json:"riskParameters"`
	SymbolType string `json:"symbolType"`
}

type InstrumentInfoLinear struct {
	Symbol          string `json:"symbol"`
	ContractType    string `json:"contractType"`
	Status          string `json:"status"`
	BaseCoin        string `json:"baseCoin"`
	QuoteCoin       string `json:"quoteCoin"`
	SymbolType      string `json:"symbolType"`
	LaunchTime      string `json:"launchTime"`
	DeliveryTime    string `json:"deliveryTime"`
	DeliveryFeeRate string `json:"deliveryFeeRate"`
	PriceScale      string `json:"priceScale"`
	LeverageFilter  struct {
		MinLeverage  string `json:"minLeverage"`
		MaxLeverage  string `json:"maxLeverage"`
		LeverageStep string `json:"leverageStep"`
	} `json:"leverageFilter"`
	PriceFilter struct {
		MinPrice string `json:"minPrice"`
		MaxPrice string `json:"maxPrice"`
		TickSize string `json:"tickSize"`
	} `json:"priceFilter"`
	LotSizeFilter struct {
		MaxOrderQty         string `json:"maxOrderQty"`
		MinOrderQty         string `json:"minOrderQty"`
		QtyStep             string `json:"qtyStep"`
		PostOnlyMaxOrderQty string `json:"postOnlyMaxOrderQty"`
		MaxMktOrderQty      string `json:"maxMktOrderQty"`
		MinNotionalValue    string `json:"minNotionalValue"`
	} `json:"lotSizeFilter"`
	UnifiedMarginTrade bool            `json:"unifiedMarginTrade"`
	FundingInterval    int             `json:"fundingInterval"`
	SettleCoin         string          `json:"settleCoin"`
	CopyTrading        string          `json:"copyTrading"`
	UpperFundingRate   string          `json:"upperFundingRate"`
	LowerFundingRate   string          `json:"lowerFundingRate"`
	IsPreListing       bool            `json:"isPreListing"`
	PreListingInfo     *PreListingInfo `json:"preListingInfo,omitempty"`
	RiskParameters     struct {
		PriceLimitRatioX string `json:"priceLimitRatioX"`
		PriceLimitRatioY string `json:"priceLimitRatioY"`
	} `json:"riskParameters"`
	ForbidUplWithdrawal bool   `json:"forbidUplWithdrawal,omitempty"`
	DisplayName         string `json:"displayName,omitempty"`
}

type PreListingInfo struct {
	CurAuctionPhase string `json:"curAuctionPhase"`
	Phases          []struct {
		Phase     string `json:"phase"`
		StartTime string `json:"startTime"`
		EndTime   string `json:"endTime"`
	} `json:"phases"`
	AuctionFeeInfo struct {
		AuctionFeeRate string `json:"auctionFeeRate"`
		TakerFeeRate   string `json:"takerFeeRate"`
		MakerFeeRate   string `json:"makerFeeRate"`
	} `json:"auctionFeeInfo"`
	SkipCallAuction bool `json:"skipCallAuction"`
}

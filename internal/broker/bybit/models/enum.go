package models

type Category string

const (
	CategoryDefault Category = ""
	CategorySpot    Category = "spot"
	CategoryLinear  Category = "linear"
	CategoryInverse Category = "inverse"
)

type Interval string

const (
	Interval1Min   Interval = "1"
	Interval3Min   Interval = "3"
	Interval5Min   Interval = "5"
	Interval15Min  Interval = "15"
	Interval30Min  Interval = "30"
	Interval60Min  Interval = "60"
	Interval120Min Interval = "120"
	Interval240Min Interval = "240"
	Interval360Min Interval = "360"
	Interval720Min Interval = "720"
	Interval1Day   Interval = "D"
	Interval1Week  Interval = "W"
	Interval1Month Interval = "M"
)

type Side string

const (
	SideBuy  Side = "Buy"
	SideSell Side = "Sell"
)

type OrderType string

const (
	OrderTypeMarket OrderType = "Market"
	OrderTypeLimit  OrderType = "Limit"
)

type MarketUnit string

const (
	MarketUnitBaseCoin  MarketUnit = "baseCoin"  // for buy orders
	MarketUnitQuoteCoin MarketUnit = "quoteCoin" // for sell orders
)

type SlippageToleranceType string

const (
	SlippageToleranceTypeTickSize SlippageToleranceType = "TickSize"
	SlippageToleranceTypePercent  SlippageToleranceType = "Percent"
)

type TriggerDirection int

const (
	TriggerDirectionRise TriggerDirection = 1 // triggered when market price rises to triggerPrice
	TriggerDirectionFall TriggerDirection = 2 // triggered when market price falls to triggerPrice
)

type OrderFilter string

const (
	OrderFilterOrder     OrderFilter = "Order"
	OrderFilterTpslOrder OrderFilter = "tpslOrder" // Spot TP/SL order
	OrderFilterStopOrder OrderFilter = "StopOrder" // Spot conditional order
)

type TriggerBy string

const (
	TriggerByLastPrice  TriggerBy = "LastPrice"
	TriggerByIndexPrice TriggerBy = "IndexPrice"
	TriggerByMarkPrice  TriggerBy = "MarkPrice"
	// TriggerByPrevPrice  TriggerBy = "PrevPrice"
)

type TimeInForce string

const (
	TimeInForceGTC TimeInForce = "GTC" // Good Till Cancel
	TimeInForceIOC TimeInForce = "IOC" // Immediate or Cancel
	// Market orders always use IOC
)

type PositionIdx int

const (
	PositionIdxOneWay    PositionIdx = 0 // one-way mode
	PositionIdxHedgeBuy  PositionIdx = 1 // hedge-mode Buy side
	PositionIdxHedgeSell PositionIdx = 2 // hedge-mode Sell side
)

type SMPType string

const (
	SMPTypeNone        SMPType = "None"
	SMPTypeCancelMaker SMPType = "CancelMaker"
	SMPTypeCancelTaker SMPType = "CancelTaker"
	SMPTypeCancelBoth  SMPType = "CancelBoth"
)

type TpslMode string

const (
	TpslModeFull    TpslMode = "Full"    // entire position for TP/SL
	TpslModePartial TpslMode = "Partial" // partial position tp/sl
	TpslModeSpot    TpslMode = "Spot"
)

type BboSideType string

const (
	BboSideTypeQueue        BboSideType = "Queue"        // use order price on orderbook in same direction as side
	BboSideTypeCounterparty BboSideType = "Counterparty" // use order price on orderbook in opposite direction as side
)

// CancelTakeProfitValue
const CancelTakeProfitValue = "0"

// CancelStopLossValue
const CancelStopLossValue = "0"

type OrderStatus string

const (
	OrderStatusNew                     OrderStatus = "New"
	OrderStatusPartiallyFilled         OrderStatus = "PartiallyFilled"
	OrderStatusUntriggered             OrderStatus = "Untriggered"
	OrderStatusActive                  OrderStatus = "Active"
	OrderStatusCreated                 OrderStatus = "Created"
	OrderStatusRejected                OrderStatus = "Rejected"
	OrderStatusPartiallyFilledCanceled OrderStatus = "PartiallyFilledCanceled"
	OrderStatusFilled                  OrderStatus = "Filled"
	OrderStatusCancelled               OrderStatus = "Cancelled"
	OrderStatusTriggered               OrderStatus = "Triggered"
	OrderStatusDeactivated             OrderStatus = "Deactivated"
)

type StopOrderType string

const (
	StopOrderTypeTakeProfit             StopOrderType = "TakeProfit"
	StopOrderTypeStopLoss               StopOrderType = "StopLoss"
	StopOrderTypeTrailingStop           StopOrderType = "TrailingStop"
	StopOrderTypeStop                   StopOrderType = "Stop"
	StopOrderTypePartialTakeProfit      StopOrderType = "PartialTakeProfit"
	StopOrderTypePartialStopLoss        StopOrderType = "PartialStopLoss"
	StopOrderTypeTpslOrder              StopOrderType = "tpslOrder"
	StopOrderTypeOcoOrder               StopOrderType = "OcoOrder"
	StopOrderTypeMmRateClose            StopOrderType = "MmRateClose"
	StopOrderTypeBidirectionalTpslOrder StopOrderType = "BidirectionalTpslOrder"
)

type PositionStatus string

const (
	PositionStatusNormal PositionStatus = "Normal"
	PositionStatusLiq    PositionStatus = "Liq"
	PositionStatusAdl    PositionStatus = "Adl"
)

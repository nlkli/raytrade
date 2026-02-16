package bybit

import (
	"errors"
	"nlkli/raytrade/internal/broker/bybit/models"
)

// https://bybit-exchange.github.io/docs/v5/order/create-order
func (c *Client) PlaceOrder(
	// ctx context.Context,
	category models.Category,
	symbol string,
	side models.Side,
	orderType models.OrderType,
	qty string,
	opts ...OrderOption,
) error {
	o := defaultOrderOptions()
	for _, opt := range opts {
		opt(o)
	}

	// ... implementation using o ...
	return nil
}

// OrderOptions holds all optional parameters for PlaceOrder
type OrderOptions struct {
	IsLeverage            *int
	MarketUnit            *models.MarketUnit
	SlippageToleranceType *models.SlippageToleranceType
	SlippageTolerance     *string
	Price                 *string
	TriggerDirection      *models.TriggerDirection
	OrderFilter           *models.OrderFilter
	TriggerPrice          *string
	TriggerBy             *models.TriggerBy
	OrderIv               *string
	TimeInForce           *models.TimeInForce
	PositionIdx           *models.PositionIdx
	OrderLinkId           *string
	TakeProfit            *string
	StopLoss              *string
	TpTriggerBy           *models.TpSlTriggerBy
	SlTriggerBy           *models.TpSlTriggerBy
	ReduceOnly            *bool
	CloseOnTrigger        *bool
	SmpType               *string
	Mmp                   *bool
	TpslMode              *models.TpslMode
	TpLimitPrice          *string
	SlLimitPrice          *string
	TpOrderType           *models.OrderType
	SlOrderType           *models.OrderType
	BboSideType           *models.BboSideType
	BboLevel              *string
}

// OrderOption defines the function signature for order options
type OrderOption func(*OrderOptions)

func defaultOrderOptions() *OrderOptions {
	return &OrderOptions{}
}

// Whether to borrow.
// 0 (default): false, spot trading
// 1: true, margin trading, make sure you turn on margin trading, and set the relevant currency as collateral
func WithIsLeverage(leverage int) OrderOption {
	return func(o *OrderOptions) {
		o.IsLeverage = &leverage
	}
}

// Select the unit for qty when create Spot market orders
// baseCoin: for example, buy BTCUSDT, then "qty" unit is BTC
// quoteCoin: for example, sell BTCUSDT, then "qty" unit is USDT
func WithMarketUnit(unit models.MarketUnit) OrderOption {
	return func(o *OrderOptions) {
		o.MarketUnit = &unit
	}
}

// Slippage tolerance Type for market order, TickSize, Percent
// take profit, stoploss, conditional orders are not supported
// TickSize:
// the highest price of Buy order = ask1 + slippageTolerance x tickSize;
// the lowest price of Sell order = bid1 - slippageTolerance x tickSize
// Percent:
// the highest price of Buy order = ask1 x (1 + slippageTolerance x 0.01);
// the lowest price of Sell order = bid1 x (1 - slippageTolerance x 0.01)
func WithSlippageToleranceType(t models.SlippageToleranceType) OrderOption {
	return func(o *OrderOptions) {
		o.SlippageToleranceType = &t
	}
}

// Slippage tolerance value
// TickSize: range is [1, 10000], integer only
// Percent: range is [0.01, 10], up to 2 decimals
func WithSlippageTolerance(tolerance string) OrderOption {
	return func(o *OrderOptions) {
		o.SlippageTolerance = &tolerance
	}
}

// Order price
// Market order will ignore this field
// Please check the min price and price precision from instrument info endpoint
// If you have position, price needs to be better than liquidation price
func WithPrice(price string) OrderOption {
	return func(o *OrderOptions) {
		o.Price = &price
	}
}

// Conditional order param. Used to identify the expected direction of the conditional order.
// 1: triggered when market price rises to triggerPrice
// 2: triggered when market price falls to triggerPrice
// Valid for linear & inverse
func WithTriggerDirection(dir models.TriggerDirection) OrderOption {
	return func(o *OrderOptions) {
		o.TriggerDirection = &dir
	}
}

// If it is not passed, Order by default.
// Order
// tpslOrder: Spot TP/SL order, the assets are occupied even before the order is triggered
// StopOrder: Spot conditional order, the assets will not be occupied until the price of the underlying asset reaches the trigger price, and the required assets will be occupied after the Conditional order is triggered
// Valid for spot only
func WithOrderFilter(filter models.OrderFilter) OrderOption {
	return func(o *OrderOptions) {
		o.OrderFilter = &filter
	}
}

// For Perps & Futures, it is the conditional order trigger price. If you expect the price to rise to trigger your conditional order, make sure:
// triggerPrice > market price
// Else, triggerPrice < market price
// For spot, it is the TP/SL and Conditional order trigger price
func WithTriggerPrice(price string) OrderOption {
	return func(o *OrderOptions) {
		o.TriggerPrice = &price
	}
}

// Trigger price type, Conditional order param for Perps & Futures.
// LastPrice
// IndexPrice
// MarkPrice
// Valid for linear & inverse
func WithTriggerBy(triggerBy models.TriggerBy) OrderOption {
	return func(o *OrderOptions) {
		o.TriggerBy = &triggerBy
	}
}

// Implied volatility. option only. Pass the real value, e.g for 10%, 0.1 should be passed. orderIv has a higher priority when price is passed as well
func WithOrderIv(iv string) OrderOption {
	return func(o *OrderOptions) {
		o.OrderIv = &iv
	}
}

// Time in force
// Market order will always use IOC
// If not passed, GTC is used by default
func WithTimeInForce(tif models.TimeInForce) OrderOption {
	return func(o *OrderOptions) {
		o.TimeInForce = &tif
	}
}

// Used to identify positions in different position modes. Under hedge-mode, this param is required
// 0: one-way mode
// 1: hedge-mode Buy side
// 2: hedge-mode Sell side
func WithPositionIdx(idx models.PositionIdx) OrderOption {
	return func(o *OrderOptions) {
		o.PositionIdx = &idx
	}
}

// User customised order ID. A max of 36 characters. Combinations of numbers, letters (upper and lower cases), dashes, and underscores are supported.
// Futures, Perps & Spot: orderLinkId rules:
// optional param
// always unique
// Options orderLinkId rules:
// required param
// always unique
func WithOrderLinkId(linkId string) OrderOption {
	return func(o *OrderOptions) {
		o.OrderLinkId = &linkId
	}
}

// Take profit price
// Spot Limit order supports take profit, stop loss or limit take profit, limit stop loss when creating an order
func WithTakeProfit(tp string) OrderOption {
	return func(o *OrderOptions) {
		o.TakeProfit = &tp
	}
}

// Stop loss price
// Spot Limit order supports take profit, stop loss or limit take profit, limit stop loss when creating an order
func WithStopLoss(sl string) OrderOption {
	return func(o *OrderOptions) {
		o.StopLoss = &sl
	}
}

// The price type to trigger take profit. MarkPrice, IndexPrice, default: LastPrice. Valid for linear & inverse
func WithTpTriggerBy(tpBy models.TpSlTriggerBy) OrderOption {
	return func(o *OrderOptions) {
		o.TpTriggerBy = &tpBy
	}
}

// The price type to trigger stop loss. MarkPrice, IndexPrice, default: LastPrice. Valid for linear & inverse
func WithSlTriggerBy(slBy models.TpSlTriggerBy) OrderOption {
	return func(o *OrderOptions) {
		o.SlTriggerBy = &slBy
	}
}

// https://www.bybit.com/en/help-center/article/Reduce-Only-Order
// What is a reduce-only order? true means your position can only reduce in size if this order is triggered.
// You must specify it as true when you are about to close/reduce the position
// When reduceOnly is true, take profit/stop loss cannot be set
// Valid for linear, inverse & option
func WithReduceOnly(reduceOnly bool) OrderOption {
	return func(o *OrderOptions) {
		o.ReduceOnly = &reduceOnly
	}
}

// https://www.bybit.com/en/help-center/article/Close-On-Trigger-Order
// What is a close on trigger order? For a closing order. It can only reduce your position, not increase it. If the account has insufficient available balance when the closing order is triggered, then other active orders of similar contracts will be cancelled or reduced. It can be used to ensure your stop loss reduces your position regardless of current available margin.
// Valid for linear & inverse
func WithCloseOnTrigger(closeOnTrigger bool) OrderOption {
	return func(o *OrderOptions) {
		o.CloseOnTrigger = &closeOnTrigger
	}
}

// Smp execution type. What is SMP?
// https://bybit-exchange.github.io/docs/v5/smp
func WithSmpType(smpType string) OrderOption {
	return func(o *OrderOptions) {
		o.SmpType = &smpType
	}
}

// Market maker protection. option only. true means set the order as a market maker protection order. What is mmp?
// https://bybit-exchange.github.io/docs/v5/account/set-mmp
func WithMmp(mmp bool) OrderOption {
	return func(o *OrderOptions) {
		o.Mmp = &mmp
	}
}

// TP/SL mode
// Full: entire position for TP/SL. Then, tpOrderType or slOrderType must be Market
// Partial: partial position tp/sl (as there is no size option, so it will create tp/sl orders with the qty you actually fill). Limit TP/SL order are supported. Note: When create limit tp/sl, tpslMode is required and it must be Partial
// Valid for linear & inverse
func WithTpslMode(mode models.TpslMode) OrderOption {
	return func(o *OrderOptions) {
		o.TpslMode = &mode
	}
}

// The limit order price when take profit price is triggered
// linear & inverse: only works when tpslMode=Partial and tpOrderType=Limit
// Spot: it is required when the order has takeProfit and "tpOrderType"=Limit
func WithTpLimitPrice(price string) OrderOption {
	return func(o *OrderOptions) {
		o.TpLimitPrice = &price
	}
}

// The limit order price when stop loss price is triggered
// linear & inverse: only works when tpslMode=Partial and slOrderType=Limit
// Spot: it is required when the order has stopLoss and "slOrderType"=Limit
func WithSlLimitPrice(price string) OrderOption {
	return func(o *OrderOptions) {
		o.SlLimitPrice = &price
	}
}

// The order type when take profit is triggered
// linear & inverse: Market(default), Limit. For tpslMode=Full, it only supports tpOrderType=Market
// Spot:
// Market: when you set "takeProfit",
// Limit: when you set "takeProfit" and "tpLimitPrice"
func WithTpOrderType(orderType models.OrderType) OrderOption {
	return func(o *OrderOptions) {
		o.TpOrderType = &orderType
	}
}

// The order type when stop loss is triggered
// linear & inverse: Market(default), Limit. For tpslMode=Full, it only supports slOrderType=Market
// Spot:
// Market: when you set "stopLoss",
// Limit: when you set "stopLoss" and "slLimitPrice"
func WithSlOrderType(orderType models.OrderType) OrderOption {
	return func(o *OrderOptions) {
		o.SlOrderType = &orderType
	}
}

// Queue: use the order price on the orderbook in the same direction as the side
// Counterparty: use the order price on the orderbook in the opposite direction as the side
// Valid for linear & inverse
func WithBboSideType(sideType models.BboSideType) OrderOption {
	return func(o *OrderOptions) {
		o.BboSideType = &sideType
	}
}

// 1,2,3,4,5 Valid for linear & inverse
func WithBboLevel(level string) OrderOption {
	return func(o *OrderOptions) {
		o.BboLevel = &level
	}
}

// https://bybit-exchange.github.io/docs/v5/order/amend-order
func (c *Client) AmendOrder(
	// ctx context.Context,
	category models.Category,
	symbol string,
	opts ...AmendOrderOption,
) error {
	o := defaultAmendOrderOptions()
	for _, opt := range opts {
		opt(o)
	}

	// Validate that either orderId or orderLinkId is provided
	if o.OrderId == nil && o.OrderLinkId == nil {
		return errors.New("either orderId or orderLinkId is required")
	}

	// ... implementation using o ...
	return nil
}

// AmendOrderOptions holds all optional parameters for AmendOrder
type AmendOrderOptions struct {
	OrderId      *string
	OrderLinkId  *string
	OrderIv      *string
	TriggerPrice *string
	Qty          *string
	Price        *string
	TpslMode     *models.TpslMode
	TakeProfit   *string
	StopLoss     *string
	TpTriggerBy  *models.TpSlTriggerBy
	SlTriggerBy  *models.TpSlTriggerBy
	TriggerBy    *models.TriggerBy
	TpLimitPrice *string
	SlLimitPrice *string
}

// AmendOrderOption defines the function signature for amend order options
type AmendOrderOption func(*AmendOrderOptions)

func defaultAmendOrderOptions() *AmendOrderOptions {
	return &AmendOrderOptions{}
}

// WithAmendOrderId sets the order ID to amend
func WithAmendOrderId(orderId string) AmendOrderOption {
	return func(o *AmendOrderOptions) {
		o.OrderId = &orderId
	}
}

// WithAmendOrderLinkId sets the custom order ID to amend
func WithAmendOrderLinkId(orderLinkId string) AmendOrderOption {
	return func(o *AmendOrderOptions) {
		o.OrderLinkId = &orderLinkId
	}
}

// WithAmendOrderIv sets implied volatility for options
func WithAmendOrderIv(orderIv string) AmendOrderOption {
	return func(o *AmendOrderOptions) {
		o.OrderIv = &orderIv
	}
}

// WithAmendTriggerPrice sets new trigger price for conditional orders
func WithAmendTriggerPrice(triggerPrice string) AmendOrderOption {
	return func(o *AmendOrderOptions) {
		o.TriggerPrice = &triggerPrice
	}
}

// WithAmendQty sets new order quantity
func WithAmendQty(qty string) AmendOrderOption {
	return func(o *AmendOrderOptions) {
		o.Qty = &qty
	}
}

// WithAmendPrice sets new order price
func WithAmendPrice(price string) AmendOrderOption {
	return func(o *AmendOrderOptions) {
		o.Price = &price
	}
}

// WithAmendTpslMode sets TP/SL mode
func WithAmendTpslMode(mode models.TpslMode) AmendOrderOption {
	return func(o *AmendOrderOptions) {
		o.TpslMode = &mode
	}
}

// WithAmendTakeProfit sets new take profit price
// Pass "0" to cancel existing take profit
func WithAmendTakeProfit(takeProfit string) AmendOrderOption {
	return func(o *AmendOrderOptions) {
		o.TakeProfit = &takeProfit
	}
}

// WithAmendStopLoss sets new stop loss price
// Pass "0" to cancel existing stop loss
func WithAmendStopLoss(stopLoss string) AmendOrderOption {
	return func(o *AmendOrderOptions) {
		o.StopLoss = &stopLoss
	}
}

// WithAmendTpTriggerBy sets price type for take profit trigger
func WithAmendTpTriggerBy(tpTriggerBy models.TpSlTriggerBy) AmendOrderOption {
	return func(o *AmendOrderOptions) {
		o.TpTriggerBy = &tpTriggerBy
	}
}

// WithAmendSlTriggerBy sets price type for stop loss trigger
func WithAmendSlTriggerBy(slTriggerBy models.TpSlTriggerBy) AmendOrderOption {
	return func(o *AmendOrderOptions) {
		o.SlTriggerBy = &slTriggerBy
	}
}

// WithAmendTriggerBy sets trigger price type
func WithAmendTriggerBy(triggerBy models.TriggerBy) AmendOrderOption {
	return func(o *AmendOrderOptions) {
		o.TriggerBy = &triggerBy
	}
}

// WithAmendTpLimitPrice sets limit price when take profit triggered
func WithAmendTpLimitPrice(tpLimitPrice string) AmendOrderOption {
	return func(o *AmendOrderOptions) {
		o.TpLimitPrice = &tpLimitPrice
	}
}

// WithAmendSlLimitPrice sets limit price when stop loss triggered
func WithAmendSlLimitPrice(slLimitPrice string) AmendOrderOption {
	return func(o *AmendOrderOptions) {
		o.SlLimitPrice = &slLimitPrice
	}
}

// https://bybit-exchange.github.io/docs/v5/order/cancel-order
func (c *Client) CancelOrder(
	// ctx context.Context,
	category models.Category,
	symbol string,
	orderId *string,
	orderLinkId *string,
	orderFilter *models.OrderFilter,
) error {
	// Validate that either orderId or orderLinkId is provided
	if orderId == nil && orderLinkId == nil {
		return errors.New("either orderId or orderLinkId is required")
	}

	// ... implementation ...
	return nil
}

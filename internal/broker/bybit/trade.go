package bybit

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"nlkli/raytrade/internal/broker/bybit/models"
)

// https://bybit-exchange.github.io/docs/v5/order/create-order
func (c *Client) PlaceOrder(

	ctx context.Context,

	category models.Category,
	symbol string,
	side models.Side,
	orderType models.OrderType,
	qty string,

	opts ...PlaceOrderOption,

) (*models.OrderResult, error) {

	params := &PlaceOrderRequestParams{
		Category:  category,
		Symbol:    symbol,
		Side:      side,
		OrderType: orderType,
		Qty:       qty,
	}

	for _, opt := range opts {
		opt(params)
	}

	jsonParams, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}

	fullURL := fmt.Sprintf("%s%s", c.baseURL, "/v5/order/create")
	req, err := http.NewRequestWithContext(
		ctx, "POST", fullURL, bytes.NewReader(jsonParams),
	)
	if err != nil {
		return nil, err
	}

	var order models.OrderResult
	err = c.callAPI(req, string(jsonParams), &order)
	if err != nil {
		return nil, err
	}

	return &order, nil
}

type PlaceOrderRequestParams struct {
	// Required
	Category  models.Category  `json:"category"`
	Symbol    string           `json:"symbol"`
	Side      models.Side      `json:"side"`
	OrderType models.OrderType `json:"orderType"`
	Qty       string           `json:"qty"`

	// Options
	IsLeverage            *int                          `json:"isLeverage,omitempty"`
	MarketUnit            *models.MarketUnit            `json:"marketUnit,omitempty"`
	SlippageToleranceType *models.SlippageToleranceType `json:"slippageToleranceType,omitempty"`
	SlippageTolerance     *string                       `json:"slippageTolerance,omitempty"`
	Price                 *string                       `json:"price,omitempty"`
	TriggerDirection      *models.TriggerDirection      `json:"triggerDirection,omitempty"`
	OrderFilter           *models.OrderFilter           `json:"orderFilter,omitempty"`
	TriggerPrice          *string                       `json:"triggerPrice,omitempty"`
	TriggerBy             *models.TriggerBy             `json:"triggerBy,omitempty"`
	OrderIv               *string                       `json:"orderIv,omitempty"`
	TimeInForce           *models.TimeInForce           `json:"timeInForce,omitempty"`
	PositionIdx           *models.PositionIdx           `json:"positionIdx,omitempty"`
	OrderLinkId           *string                       `json:"orderLinkId,omitempty"`
	TakeProfit            *string                       `json:"takeProfit,omitempty"`
	StopLoss              *string                       `json:"stopLoss,omitempty"`
	TpTriggerBy           *models.TpSlTriggerBy         `json:"tpTriggerBy,omitempty"`
	SlTriggerBy           *models.TpSlTriggerBy         `json:"slTriggerBy,omitempty"`
	ReduceOnly            *bool                         `json:"reduceOnly,omitempty"`
	CloseOnTrigger        *bool                         `json:"closeOnTrigger,omitempty"`
	SmpType               *string                       `json:"smpType,omitempty"`
	Mmp                   *bool                         `json:"mmp,omitempty"`
	TpslMode              *models.TpslMode              `json:"tpslMode,omitempty"`
	TpLimitPrice          *string                       `json:"tpLimitPrice,omitempty"`
	SlLimitPrice          *string                       `json:"slLimitPrice,omitempty"`
	TpOrderType           *models.OrderType             `json:"tpOrderType,omitempty"`
	SlOrderType           *models.OrderType             `json:"slOrderType,omitempty"`
	BboSideType           *models.BboSideType           `json:"bboSideType,omitempty"`
	BboLevel              *int                          `json:"bboLevel,omitempty"`
}

// PlaceOrderOption defines the function signature for order options
type PlaceOrderOption func(*PlaceOrderRequestParams)

// Whether to borrow.
// 0 (default): false, spot trading
// 1: true, margin trading, make sure you turn on margin trading, and set the relevant currency as collateral
func WithIsLeverage(leverage int) PlaceOrderOption {
	return func(o *PlaceOrderRequestParams) {
		o.IsLeverage = &leverage
	}
}

// Select the unit for qty when create Spot market orders
// baseCoin: for example, buy BTCUSDT, then "qty" unit is BTC
// quoteCoin: for example, sell BTCUSDT, then "qty" unit is USDT
func WithMarketUnit(unit models.MarketUnit) PlaceOrderOption {
	return func(o *PlaceOrderRequestParams) {
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
func WithSlippageToleranceType(t models.SlippageToleranceType) PlaceOrderOption {
	return func(o *PlaceOrderRequestParams) {
		o.SlippageToleranceType = &t
	}
}

// Slippage tolerance value
// TickSize: range is [1, 10000], integer only
// Percent: range is [0.01, 10], up to 2 decimals
func WithSlippageTolerance(tolerance string) PlaceOrderOption {
	return func(o *PlaceOrderRequestParams) {
		o.SlippageTolerance = &tolerance
	}
}

// Order price
// Market order will ignore this field
// Please check the min price and price precision from instrument info endpoint
// If you have position, price needs to be better than liquidation price
func WithPrice(price string) PlaceOrderOption {
	return func(o *PlaceOrderRequestParams) {
		o.Price = &price
	}
}

// Conditional order param. Used to identify the expected direction of the conditional order.
// 1: triggered when market price rises to triggerPrice
// 2: triggered when market price falls to triggerPrice
// Valid for linear & inverse
func WithTriggerDirection(dir models.TriggerDirection) PlaceOrderOption {
	return func(o *PlaceOrderRequestParams) {
		o.TriggerDirection = &dir
	}
}

// If it is not passed, Order by default.
// Order
// tpslOrder: Spot TP/SL order, the assets are occupied even before the order is triggered
// StopOrder: Spot conditional order, the assets will not be occupied until the price of the underlying asset reaches the trigger price, and the required assets will be occupied after the Conditional order is triggered
// Valid for spot only
func WithOrderFilter(filter models.OrderFilter) PlaceOrderOption {
	return func(o *PlaceOrderRequestParams) {
		o.OrderFilter = &filter
	}
}

// For Perps & Futures, it is the conditional order trigger price. If you expect the price to rise to trigger your conditional order, make sure:
// triggerPrice > market price
// Else, triggerPrice < market price
// For spot, it is the TP/SL and Conditional order trigger price
func WithTriggerPrice(price string) PlaceOrderOption {
	return func(o *PlaceOrderRequestParams) {
		o.TriggerPrice = &price
	}
}

// Trigger price type, Conditional order param for Perps & Futures.
// LastPrice
// IndexPrice
// MarkPrice
// Valid for linear & inverse
func WithTriggerBy(triggerBy models.TriggerBy) PlaceOrderOption {
	return func(o *PlaceOrderRequestParams) {
		o.TriggerBy = &triggerBy
	}
}

// Implied volatility. option only. Pass the real value, e.g for 10%, 0.1 should be passed. orderIv has a higher priority when price is passed as well
func WithOrderIv(iv string) PlaceOrderOption {
	return func(o *PlaceOrderRequestParams) {
		o.OrderIv = &iv
	}
}

// Time in force
// Market order will always use IOC
// If not passed, GTC is used by default
func WithTimeInForce(tif models.TimeInForce) PlaceOrderOption {
	return func(o *PlaceOrderRequestParams) {
		o.TimeInForce = &tif
	}
}

// Used to identify positions in different position modes. Under hedge-mode, this param is required
// 0: one-way mode
// 1: hedge-mode Buy side
// 2: hedge-mode Sell side
func WithPositionIdx(idx models.PositionIdx) PlaceOrderOption {
	return func(o *PlaceOrderRequestParams) {
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
func WithOrderLinkId(linkId string) PlaceOrderOption {
	return func(o *PlaceOrderRequestParams) {
		o.OrderLinkId = &linkId
	}
}

// Take profit price
// Spot Limit order supports take profit, stop loss or limit take profit, limit stop loss when creating an order
func WithTakeProfit(tp string) PlaceOrderOption {
	return func(o *PlaceOrderRequestParams) {
		o.TakeProfit = &tp
	}
}

// Stop loss price
// Spot Limit order supports take profit, stop loss or limit take profit, limit stop loss when creating an order
func WithStopLoss(sl string) PlaceOrderOption {
	return func(o *PlaceOrderRequestParams) {
		o.StopLoss = &sl
	}
}

// The price type to trigger take profit. MarkPrice, IndexPrice, default: LastPrice. Valid for linear & inverse
func WithTpTriggerBy(tpBy models.TpSlTriggerBy) PlaceOrderOption {
	return func(o *PlaceOrderRequestParams) {
		o.TpTriggerBy = &tpBy
	}
}

// The price type to trigger stop loss. MarkPrice, IndexPrice, default: LastPrice. Valid for linear & inverse
func WithSlTriggerBy(slBy models.TpSlTriggerBy) PlaceOrderOption {
	return func(o *PlaceOrderRequestParams) {
		o.SlTriggerBy = &slBy
	}
}

// https://www.bybit.com/en/help-center/article/Reduce-Only-Order
// What is a reduce-only order? true means your position can only reduce in size if this order is triggered.
// You must specify it as true when you are about to close/reduce the position
// When reduceOnly is true, take profit/stop loss cannot be set
// Valid for linear, inverse & option
func WithReduceOnly(reduceOnly bool) PlaceOrderOption {
	return func(o *PlaceOrderRequestParams) {
		o.ReduceOnly = &reduceOnly
	}
}

// https://www.bybit.com/en/help-center/article/Close-On-Trigger-Order
// What is a close on trigger order? For a closing order. It can only reduce your position, not increase it. If the account has insufficient available balance when the closing order is triggered, then other active orders of similar contracts will be cancelled or reduced. It can be used to ensure your stop loss reduces your position regardless of current available margin.
// Valid for linear & inverse
func WithCloseOnTrigger(closeOnTrigger bool) PlaceOrderOption {
	return func(o *PlaceOrderRequestParams) {
		o.CloseOnTrigger = &closeOnTrigger
	}
}

// Smp execution type. What is SMP?
// https://bybit-exchange.github.io/docs/v5/smp
func WithSmpType(smpType string) PlaceOrderOption {
	return func(o *PlaceOrderRequestParams) {
		o.SmpType = &smpType
	}
}

// Market maker protection. option only. true means set the order as a market maker protection order. What is mmp?
// https://bybit-exchange.github.io/docs/v5/account/set-mmp
func WithMmp(mmp bool) PlaceOrderOption {
	return func(o *PlaceOrderRequestParams) {
		o.Mmp = &mmp
	}
}

// TP/SL mode
// Full: entire position for TP/SL. Then, tpOrderType or slOrderType must be Market
// Partial: partial position tp/sl (as there is no size option, so it will create tp/sl orders with the qty you actually fill). Limit TP/SL order are supported. Note: When create limit tp/sl, tpslMode is required and it must be Partial
// Valid for linear & inverse
func WithTpslMode(mode models.TpslMode) PlaceOrderOption {
	return func(o *PlaceOrderRequestParams) {
		o.TpslMode = &mode
	}
}

// The limit order price when take profit price is triggered
// linear & inverse: only works when tpslMode=Partial and tpOrderType=Limit
// Spot: it is required when the order has takeProfit and "tpOrderType"=Limit
func WithTpLimitPrice(price string) PlaceOrderOption {
	return func(o *PlaceOrderRequestParams) {
		o.TpLimitPrice = &price
	}
}

// The limit order price when stop loss price is triggered
// linear & inverse: only works when tpslMode=Partial and slOrderType=Limit
// Spot: it is required when the order has stopLoss and "slOrderType"=Limit
func WithSlLimitPrice(price string) PlaceOrderOption {
	return func(o *PlaceOrderRequestParams) {
		o.SlLimitPrice = &price
	}
}

// The order type when take profit is triggered
// linear & inverse: Market(default), Limit. For tpslMode=Full, it only supports tpOrderType=Market
// Spot:
// Market: when you set "takeProfit",
// Limit: when you set "takeProfit" and "tpLimitPrice"
func WithTpOrderType(orderType models.OrderType) PlaceOrderOption {
	return func(o *PlaceOrderRequestParams) {
		o.TpOrderType = &orderType
	}
}

// The order type when stop loss is triggered
// linear & inverse: Market(default), Limit. For tpslMode=Full, it only supports slOrderType=Market
// Spot:
// Market: when you set "stopLoss",
// Limit: when you set "stopLoss" and "slLimitPrice"
func WithSlOrderType(orderType models.OrderType) PlaceOrderOption {
	return func(o *PlaceOrderRequestParams) {
		o.SlOrderType = &orderType
	}
}

// Queue: use the order price on the orderbook in the same direction as the side
// Counterparty: use the order price on the orderbook in the opposite direction as the side
// Valid for linear & inverse
func WithBboSideType(sideType models.BboSideType) PlaceOrderOption {
	return func(o *PlaceOrderRequestParams) {
		o.BboSideType = &sideType
	}
}

// 1,2,3,4,5 Valid for linear & inverse
func WithBboLevel(level int) PlaceOrderOption {
	return func(o *PlaceOrderRequestParams) {
		o.BboLevel = &level
	}
}

// https://bybit-exchange.github.io/docs/v5/order/amend-order
func (c *Client) AmendOrder(

	ctx context.Context,

	category models.Category,
	symbol string,
	opts ...AmendOrderOption,

) (*models.OrderResult, error) {
	params := &AmendOrderRequestParams{
		Category: category,
		Symbol:   symbol,
	}

	for _, opt := range opts {
		opt(params)
	}

	// Validate that either orderId or orderLinkId is provided
	if params.OrderId == nil && params.OrderLinkId == nil {
		return nil, errors.New("either orderId or orderLinkId is required")
	}

	jsonParams, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}

	fullURL := fmt.Sprintf("%s%s", c.baseURL, "/v5/order/amend")
	req, err := http.NewRequestWithContext(
		ctx, "POST", fullURL, bytes.NewReader(jsonParams),
	)
	if err != nil {
		return nil, err
	}

	var order models.OrderResult
	err = c.callAPI(req, string(jsonParams), &order)
	if err != nil {
		return nil, err
	}

	return &order, nil
}

type AmendOrderRequestParams struct {
	// Required
	Category models.Category `json:"category"`
	Symbol   string          `json:"symbol"`

	// Options
	OrderId      *string               `json:"orderId,omitempty"`
	OrderLinkId  *string               `json:"orderLinkId,omitempty"`
	OrderIv      *string               `json:"orderIv,omitempty"`
	TriggerPrice *string               `json:"triggerPrice,omitempty"`
	Qty          *string               `json:"qty,omitempty"`
	Price        *string               `json:"price,omitempty"`
	TpslMode     *models.TpslMode      `json:"tpslMode,omitempty"`
	TakeProfit   *string               `json:"takeProfit,omitempty"`
	StopLoss     *string               `json:"stopLoss,omitempty"`
	TpTriggerBy  *models.TpSlTriggerBy `json:"tpTriggerBy,omitempty"`
	SlTriggerBy  *models.TpSlTriggerBy `json:"slTriggerBy,omitempty"`
	TriggerBy    *models.TriggerBy     `json:"triggerBy,omitempty"`
	TpLimitPrice *string               `json:"tpLimitPrice,omitempty"`
	SlLimitPrice *string               `json:"slLimitPrice,omitempty"`
}

// AmendOrderOption defines the function signature for amend order options
type AmendOrderOption func(*AmendOrderRequestParams)

// WithAmendOrderId sets the order ID to amend
func WithAmendOrderId(orderId string) AmendOrderOption {
	return func(p *AmendOrderRequestParams) {
		p.OrderId = &orderId
	}
}

// WithAmendOrderLinkId sets the custom order ID to amend
func WithAmendOrderLinkId(orderLinkId string) AmendOrderOption {
	return func(p *AmendOrderRequestParams) {
		p.OrderLinkId = &orderLinkId
	}
}

// WithAmendOrderIv sets implied volatility for options
func WithAmendOrderIv(orderIv string) AmendOrderOption {
	return func(p *AmendOrderRequestParams) {
		p.OrderIv = &orderIv
	}
}

// WithAmendTriggerPrice sets new trigger price for conditional orders
func WithAmendTriggerPrice(triggerPrice string) AmendOrderOption {
	return func(p *AmendOrderRequestParams) {
		p.TriggerPrice = &triggerPrice
	}
}

// WithAmendQty sets new order quantity
func WithAmendQty(qty string) AmendOrderOption {
	return func(p *AmendOrderRequestParams) {
		p.Qty = &qty
	}
}

// WithAmendPrice sets new order price
func WithAmendPrice(price string) AmendOrderOption {
	return func(p *AmendOrderRequestParams) {
		p.Price = &price
	}
}

// WithAmendTpslMode sets TP/SL mode
func WithAmendTpslMode(mode models.TpslMode) AmendOrderOption {
	return func(p *AmendOrderRequestParams) {
		p.TpslMode = &mode
	}
}

// WithAmendTakeProfit sets new take profit price
// Pass "0" to cancel existing take profit
func WithAmendTakeProfit(takeProfit string) AmendOrderOption {
	return func(p *AmendOrderRequestParams) {
		p.TakeProfit = &takeProfit
	}
}

// WithAmendStopLoss sets new stop loss price
// Pass "0" to cancel existing stop loss
func WithAmendStopLoss(stopLoss string) AmendOrderOption {
	return func(p *AmendOrderRequestParams) {
		p.StopLoss = &stopLoss
	}
}

// WithAmendTpTriggerBy sets price type for take profit trigger
func WithAmendTpTriggerBy(tpTriggerBy models.TpSlTriggerBy) AmendOrderOption {
	return func(p *AmendOrderRequestParams) {
		p.TpTriggerBy = &tpTriggerBy
	}
}

// WithAmendSlTriggerBy sets price type for stop loss trigger
func WithAmendSlTriggerBy(slTriggerBy models.TpSlTriggerBy) AmendOrderOption {
	return func(p *AmendOrderRequestParams) {
		p.SlTriggerBy = &slTriggerBy
	}
}

// WithAmendTriggerBy sets trigger price type
func WithAmendTriggerBy(triggerBy models.TriggerBy) AmendOrderOption {
	return func(p *AmendOrderRequestParams) {
		p.TriggerBy = &triggerBy
	}
}

// WithAmendTpLimitPrice sets limit price when take profit triggered
func WithAmendTpLimitPrice(tpLimitPrice string) AmendOrderOption {
	return func(p *AmendOrderRequestParams) {
		p.TpLimitPrice = &tpLimitPrice
	}
}

// WithAmendSlLimitPrice sets limit price when stop loss triggered
func WithAmendSlLimitPrice(slLimitPrice string) AmendOrderOption {
	return func(p *AmendOrderRequestParams) {
		p.SlLimitPrice = &slLimitPrice
	}
}

// https://bybit-exchange.github.io/docs/v5/order/cancel-order
func (c *Client) CancelOrder(

	ctx context.Context,

	category models.Category,
	symbol string,

	opts ...CancelOrderOption,

) (*models.OrderResult, error) {

	params := &CancelOrderRequestParams{
		Category: category,
		Symbol:   symbol,
	}

	for _, opt := range opts {
		opt(params)
	}

	// Validate that either orderId or orderLinkId is provided
	if params.OrderId == nil && params.OrderLinkId == nil {
		return nil, errors.New("either orderId or orderLinkId is required")
	}

	jsonParams, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}

	fullURL := fmt.Sprintf("%s%s", c.baseURL, "/v5/order/cancel")
	req, err := http.NewRequestWithContext(
		ctx, "POST", fullURL, bytes.NewReader(jsonParams),
	)
	if err != nil {
		return nil, err
	}

	var order models.OrderResult
	err = c.callAPI(req, string(jsonParams), &order)
	if err != nil {
		return nil, err
	}

	return &order, nil
}

type CancelOrderRequestParams struct {
	// Required
	Category models.Category `json:"category"`
	Symbol   string          `json:"symbol"`

	// Options
	OrderId     *string             `json:"orderId,omitempty"`
	OrderLinkId *string             `json:"orderLinkId,omitempty"`
	OrderFilter *models.OrderFilter `json:"orderFilter,omitempty"`
}

type CancelOrderOption func(*CancelOrderRequestParams)

func WithCancelOrderID(id string) CancelOrderOption {
	return func(r *CancelOrderRequestParams) {
		r.OrderId = &id
	}
}

func WithCancelOrderLinkID(id string) CancelOrderOption {
	return func(r *CancelOrderRequestParams) {
		r.OrderLinkId = &id
	}
}

func WithCancelOrderFilter(filter models.OrderFilter) CancelOrderOption {
	return func(r *CancelOrderRequestParams) {
		r.OrderFilter = &filter
	}
}

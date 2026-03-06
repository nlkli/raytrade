package bybit

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"nlkli/raytrade/internal/broker/bybit/models"
	"strconv"
)

const MAX_ORDERLIST_LIMIT = 50

// https://bybit-exchange.github.io/docs/v5/order/realtime
func (c *Client) GetOrderList(

	ctx context.Context,
	params *GetOrderListRequestParams,

) (*models.OrderListResult, error) {

	query := make(url.Values)
	query.Set("category", string(params.Category))

	if params != nil {
		if params.Symbol != nil {
			query.Set("symbol", *params.Symbol)
		}
		if params.BaseCoin != nil {
			query.Set("baseCoin", *params.BaseCoin)
		}
		if params.SettleCoin != nil {
			query.Set("settleCoin", *params.SettleCoin)
		}
		if params.OrderId != nil {
			query.Set("orderId", *params.OrderId)
		}
		if params.OrderLinkId != nil {
			query.Set("orderLinkId", *params.OrderLinkId)
		}
		if params.OpenOnly != nil {
			query.Set("openOnly", strconv.Itoa(*params.OpenOnly))
		}
		if params.OrderFilter != nil {
			query.Set("orderFilter", string(*params.OrderFilter))
		}
		if params.Limit != nil {
			query.Set("limit", strconv.Itoa(min(max(*params.Limit, 1), MAX_ORDERLIST_LIMIT)))
		}
		if params.Cursor != nil {
			query.Set("cursor", *params.Cursor)
		}
	}

	queryString := query.Encode()
	fullURL := fmt.Sprintf("%s%s?%s", c.baseURL, "/v5/order/realtime", queryString)
	req, err := http.NewRequestWithContext(ctx, "GET", fullURL, nil)
	if err != nil {
		return nil, err
	}

	var orderListRes models.OrderListResult
	err = c.callAPI(req, queryString, &orderListRes)
	if err != nil {
		return nil, err
	}

	return &orderListRes, nil
}

// https://bybit-exchange.github.io/docs/v5/order/realtime
type GetOrderListRequestParams struct {
	// Required
	Category models.Category `json:"category"`

	// Options
	Symbol      *string
	BaseCoin    *string
	SettleCoin  *string
	OrderId     *string
	OrderLinkId *string
	OpenOnly    *int
	OrderFilter *models.OrderFilter
	Limit       *int
	Cursor      *string
}

// https://bybit-exchange.github.io/docs/v5/order/create-order
func (c *Client) PlaceOrder(

	ctx context.Context,
	params *PlaceOrderRequestParams,

) (*models.OrderResult, error) {

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

// https://bybit-exchange.github.io/docs/v5/order/create-order
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
	TpTriggerBy           *models.TriggerBy             `json:"tpTriggerBy,omitempty"`
	SlTriggerBy           *models.TriggerBy             `json:"slTriggerBy,omitempty"`
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

// https://bybit-exchange.github.io/docs/v5/order/amend-order
func (c *Client) AmendOrder(

	ctx context.Context,
	params *AmendOrderRequestParams,

) (*models.OrderResult, error) {

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

// https://bybit-exchange.github.io/docs/v5/order/amend-order
type AmendOrderRequestParams struct {
	// Required
	Category models.Category `json:"category"`
	Symbol   string          `json:"symbol"`

	// Options
	OrderId      *string           `json:"orderId,omitempty"`
	OrderLinkId  *string           `json:"orderLinkId,omitempty"`
	OrderIv      *string           `json:"orderIv,omitempty"`
	TriggerPrice *string           `json:"triggerPrice,omitempty"`
	Qty          *string           `json:"qty,omitempty"`
	Price        *string           `json:"price,omitempty"`
	TpslMode     *models.TpslMode  `json:"tpslMode,omitempty"`
	TakeProfit   *string           `json:"takeProfit,omitempty"`
	StopLoss     *string           `json:"stopLoss,omitempty"`
	TpTriggerBy  *models.TriggerBy `json:"tpTriggerBy,omitempty"`
	SlTriggerBy  *models.TriggerBy `json:"slTriggerBy,omitempty"`
	TriggerBy    *models.TriggerBy `json:"triggerBy,omitempty"`
	TpLimitPrice *string           `json:"tpLimitPrice,omitempty"`
	SlLimitPrice *string           `json:"slLimitPrice,omitempty"`
}

// https://bybit-exchange.github.io/docs/v5/order/cancel-order
func (c *Client) CancelOrder(

	ctx context.Context,
	params *CancelOrderRequestParams,

) (*models.OrderResult, error) {

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

// https://bybit-exchange.github.io/docs/v5/order/cancel-order
type CancelOrderRequestParams struct {
	// Required
	Category models.Category `json:"category"`
	Symbol   string          `json:"symbol"`

	// Options
	OrderId     *string             `json:"orderId,omitempty"`
	OrderLinkId *string             `json:"orderLinkId,omitempty"`
	OrderFilter *models.OrderFilter `json:"orderFilter,omitempty"`
}

// https://bybit-exchange.github.io/docs/v5/order/cancel-all
func (c *Client) CancelAllOrders(

	ctx context.Context,
	params *CancelAllOrdersRequestParams,

) (*models.CancelAllOrdersResult, error) {

	jsonParams, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}

	fullURL := fmt.Sprintf("%s%s", c.baseURL, "/v5/order/cancel-all")
	req, err := http.NewRequestWithContext(ctx, "POST", fullURL, bytes.NewReader(jsonParams))
	if err != nil {
		return nil, err
	}

	var cancelAllRes models.CancelAllOrdersResult
	err = c.callAPI(req, string(jsonParams), &cancelAllRes)
	if err != nil {
		return nil, err
	}

	return &cancelAllRes, nil
}

// https://bybit-exchange.github.io/docs/v5/order/cancel-all
type CancelAllOrdersRequestParams struct {
	// Required
	Category models.Category `json:"category"`

	// Options
	Symbol        *string             `json:"symbol,omitempty"`
	BaseCoin      *string             `json:"baseCoin,omitempty"`
	SettleCoin    *string             `json:"settleCoin,omitempty"`
	OrderFilter   *models.OrderFilter `json:"orderFilter,omitempty"`
	StopOrderType *string             `json:"stopOrderType,omitempty"`
}

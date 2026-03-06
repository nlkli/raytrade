package bybit

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"nlkli/raytrade/internal/broker/bybit/models"
	"strconv"
)

const (
	MAX_POSITIONINFO_LIMIT = 200
)

// https://bybit-exchange.github.io/docs/v5/position
func (c *Client) GetPositionInfo(

	ctx context.Context,

	category models.Category,
	symbol *string,
	limit *int,
	cursor *string,

) (*models.PositionInfoResult, error) {

	query := make(url.Values)
	query.Set("category", string(category))

	if symbol != nil {
		query.Set("symbol", *symbol)
	}

	if limit != nil {
		query.Set("limit", strconv.Itoa(min(max(*limit, 1), MAX_POSITIONINFO_LIMIT)))
	}

	if cursor != nil {
		query.Set("cursor", *cursor)
	}

	queryString := query.Encode()
	fullURL := fmt.Sprintf("%s%s?%s", c.baseURL, "/v5/position/list", queryString)

	req, err := http.NewRequestWithContext(ctx, "GET", fullURL, nil)
	if err != nil {
		return nil, err
	}

	var positionInfoRes models.PositionInfoResult
	err = c.callAPI(req, queryString, &positionInfoRes)
	if err != nil {
		return nil, err
	}

	return &positionInfoRes, nil
}

func (c *Client) SetTradingStop(

	ctx context.Context,
	params *TradingStopRequestParams,

) error {

	jsonParams, err := json.Marshal(params)
	if err != nil {
		return err
	}

	fullURL := fmt.Sprintf("%s%s", c.baseURL, "/v5/position/trading-stop")
	req, err := http.NewRequestWithContext(
		ctx, "POST", fullURL, bytes.NewReader(jsonParams),
	)
	if err != nil {
		return err
	}

	return c.callAPI(req, string(jsonParams), nil)
}

type TradingStopRequestParams struct {
	// Required
	Category    models.Category    `json:"category"`
	Symbol      string             `json:"symbol"`
	TpslMode    models.TpslMode    `json:"tpslMode"`
	PositionIdx models.PositionIdx `json:"positionIdx"`

	// Optional
	TakeProfit   *string           `json:"takeProfit,omitempty"`
	StopLoss     *string           `json:"stopLoss,omitempty"`
	TrailingStop *string           `json:"trailingStop,omitempty"`
	TpTriggerBy  *models.TriggerBy `json:"tpTriggerBy,omitempty"`
	SlTriggerBy  *models.TriggerBy `json:"slTriggerBy,omitempty"`
	ActivePrice  *string           `json:"activePrice,omitempty"`
	TpSize       *string           `json:"tpSize,omitempty"`
	SlSize       *string           `json:"slSize,omitempty"`
	TpLimitPrice *string           `json:"tpLimitPrice,omitempty"`
	SlLimitPrice *string           `json:"slLimitPrice,omitempty"`
	TpOrderType  *models.OrderType `json:"tpOrderType,omitempty"`
	SlOrderType  *models.OrderType `json:"slOrderType,omitempty"`
}

package bybit

import (
	"context"
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

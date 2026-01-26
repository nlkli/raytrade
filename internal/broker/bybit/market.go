package bybit

import (
	"fmt"
	"net/http"
	"net/url"
	"nlkli/raytrade/internal/broker/bybit/models"
	"strconv"
)

// https://bybit-exchange.github.io/docs/v5/market/kline
func (c *Client) GetKline(category models.Category, symbol string, interval models.Interval, start, end, limit *int) (*models.Kline, error) {
	query := make(url.Values)
	if category != models.CategoryNone {
		query.Set("category", string(category))
	}
	query.Set("symbol", symbol)
	query.Set("interval", string(interval))
	if start != nil {
		query.Set("start", strconv.Itoa(*start))
	}
	if end != nil {
		query.Set("end", strconv.Itoa(*end))
	}
	if limit != nil {
		query.Set("end", strconv.Itoa(min(max(*limit, 1), 1000)))
	}
	queryString := query.Encode()
	path := fmt.Sprintf("%s%s?%s", c.baseURL, "/v5/market/kline", queryString)
	req, err := http.NewRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}
	var kline models.Kline
	err = c.callAPI(req, queryString, &kline)
	if err != nil {
		return nil, err
	}
	return &kline, nil
}

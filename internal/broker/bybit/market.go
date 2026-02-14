package bybit

import (
	"fmt"
	"net/http"
	"net/url"
	"nlkli/raytrade/internal/broker/bybit/models"
	"strconv"
)

const (
	MAX_KLINE_LIMIT          = 1000
	MAX_SPOT_ORDERBOOK_LIMIT = 200
	MAX_ANY_ORDERBOOK_LIMIT  = 500
)

// https://bybit-exchange.github.io/docs/v5/market/kline
func (c *Client) GetKline(

	category models.Category,
	symbol string,
	interval models.Interval,
	start,
	end,
	limit *int,

) (*models.KlineResult, error) {

	query := make(url.Values)
	if category != models.CategoryDefault {
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
		query.Set("limit", strconv.Itoa(min(max(*limit, 1), MAX_KLINE_LIMIT)))
	}

	queryString := query.Encode()
	fullURL := fmt.Sprintf("%s%s?%s", c.baseURL, "/v5/market/kline", queryString)

	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return nil, err
	}

	var kline models.KlineResult
	err = c.callAPI(req, queryString, &kline)
	if err != nil {
		return nil, err
	}

	return &kline, nil
}

// https://bybit-exchange.github.io/docs/v5/market/orderbook
func (c *Client) GetOrderBook(

	category models.Category,
	symbol string,
	limit *int,

) (*models.OrderBookResult, error) {

	query := make(url.Values)
	if category != models.CategoryDefault {
		query.Set("category", string(category))
	}

	query.Set("symbol", symbol)

	if limit != nil {
		if category == models.CategorySpot {
			query.Set("limit", strconv.Itoa(min(max(*limit, 1), MAX_SPOT_ORDERBOOK_LIMIT)))
		} else {
			query.Set("limit", strconv.Itoa(min(max(*limit, 1), MAX_ANY_ORDERBOOK_LIMIT)))
		}
	}

	queryString := query.Encode()
	fullURL := fmt.Sprintf("%s%s?%s", c.baseURL, "/v5/market/orderbook", queryString)

	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return nil, err
	}

	var orderBook models.OrderBookResult
	err = c.callAPI(req, queryString, &orderBook)
	if err != nil {
		return nil, err
	}

	return &orderBook, nil
}

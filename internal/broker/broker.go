package broker

import "nlkli/raytrade/internal/cdl"

type Category int

const (
	Spot Category = iota
	Futures
)

type Broker interface {
	CandleStream(
		done <-chan struct{},
		category Category,
		symbol string,
		interval cdl.Interval,
	) (<-chan cdl.CandleStreamData, error)

	GetCandles(
		category Category,
		symbol string,
		interval cdl.Interval,
		limit int,
		start *int,
		end *int,
	) ([]cdl.Candle, error)

	// Create new slice
	ExtendStartCandles(
		candles []cdl.Candle,
		category Category,
		symbol string,
		interval cdl.Interval,
		limit int,
	) ([]cdl.Candle, error)

	// Create new slice
	ExtendEndCandles(
		candles []cdl.Candle,
		category Category,
		symbol string,
		interval cdl.Interval,
		limit int,
	) ([]cdl.Candle, error)
}

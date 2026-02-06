package broker

import "nlkli/raytrade/internal/cdl"

type Category int

const (
	Spot Category = iota
	Futures
)

type Broker interface {
	GetCandles(c Category, s string, i cdl.Interval, l int) ([]cdl.Candle, error)
}

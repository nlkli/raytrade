package cdl

type CandleArg int

const (
	Open CandleArg = iota
	High
	Low
	Close
	Volume
)

type Candle struct {
	Time   int64
	O      float64
	H      float64
	L      float64
	C      float64
	Volume float64
}

type CandleStreamData struct {
	Candle   Candle
	Interval Interval
	Confirm  bool
}

func (c *Candle) Arg(a CandleArg) float64 {
	switch a {
	case Open:
		return c.O
	case High:
		return c.H
	case Low:
		return c.L
	case Close:
		return c.C
	case Volume:
		return c.Volume
	}
	return 0
}

func MinMaxPrice(candles []Candle) (float64, float64) {
	if len(candles) == 0 {
		return 0, 0
	}
	minP, maxP := candles[0].L, candles[0].H
	for _, c := range candles[1:] {
		if c.L < minP {
			minP = c.L
		}
		if c.H < maxP {
			maxP = c.H
		}
	}
	return minP, maxP
}

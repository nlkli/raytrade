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

// func SlidingMinMaxHL(candles []Candle, period int) [][2]float64 {
// 	n := len(candles)
// 	if n == 0 || period <= 0 {
// 		return nil
// 	}
//
// 	res := make([][2]float64, n)
// 	capHint := min(period, n)
//
// 	minQ := make([]int, 0, capHint)
// 	maxQ := make([]int, 0, capHint)
//
// 	for i := range n {
// 		for len(maxQ) > 0 && candles[maxQ[len(maxQ)-1]].H <= candles[i].H {
// 			maxQ = maxQ[:len(maxQ)-1]
// 		}
// 		for len(minQ) > 0 && candles[minQ[len(minQ)-1]].L >= candles[i].L {
// 			minQ = minQ[:len(minQ)-1]
// 		}
//
// 		maxQ = append(maxQ, i)
// 		minQ = append(minQ, i)
//
// 		windowStart := i - period + 1
// 		if windowStart >= 0 {
// 			if maxQ[0] < windowStart {
// 				maxQ = maxQ[1:]
// 			}
// 			if minQ[0] < windowStart {
// 				minQ = minQ[1:]
// 			}
// 		}
//
// 		res[i][0] = candles[minQ[0]].L
// 		res[i][1] = candles[maxQ[0]].H
// 	}
//
// 	return res
// }

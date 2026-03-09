package ta

import "nlkli/raytrade/internal/cdl"

func CMA(s []float64) []float64 {
	n := len(s)
	if n == 0 {
		return []float64{}
	}
	res := make([]float64, n)
	var sum float64
	for i := range n {
		sum += s[i]
		res[i] = sum / float64(i+1)
	}
	return res
}

type baseMA struct {
	Res    []float64
	Len    int
	Period int
}

func (b *baseMA) Last() float64 {
	if b.Len == 0 {
		return 0
	}
	return b.Res[b.Len-1]
}

func (b *baseMA) MaRes() []float64 {
	return b.Res
}

func (b *baseMA) Crop() {
	if b.Len <= b.Period {
		return
	}
	start := b.Len - b.Period
	b.Res = append([]float64(nil), b.Res[start:]...)
	b.Len = len(b.Res)
}

type SMA struct {
	baseMA
	sum       float64
	CandleArg cdl.CandleArg
}

func (a *SMA) Next(candles []cdl.Candle) {
	n := len(candles)
	idx := n - 1
	oldIdx := n - min(n, a.Period+1)

	a.sum += candles[idx].Arg(a.CandleArg) - candles[oldIdx].Arg(a.CandleArg)
	a.Res = append(a.Res, a.sum/float64(a.Period))
	a.Len++
}

func NewSMA(candles []cdl.Candle, arg cdl.CandleArg, period int) *SMA {
	n := len(candles)
	if n == 0 || period <= 0 {
		return nil
	}

	res := make([]float64, n)
	var sum float64

	for i := 0; i < period && i < n; i++ {
		sum += candles[i].Arg(arg)
		res[i] = sum / float64(i+1)
	}

	for i := period; i < n; i++ {
		sum += candles[i].Arg(arg) - candles[i-period].Arg(arg)
		res[i] = sum / float64(period)
	}

	return &SMA{
		baseMA:    baseMA{Res: res, Len: n, Period: period},
		sum:       sum,
		CandleArg: arg,
	}
}

type EMA struct {
	baseMA
	alpha     float64
	CandleArg cdl.CandleArg
}

func (e *EMA) Next(candles []cdl.Candle) {
	n := len(candles)
	last := e.Res[e.Len-1]
	e.Res = append(e.Res, candles[n-1].Arg(e.CandleArg)*e.alpha+last*(1-e.alpha))
	e.Len++
}

func NewEMA(candles []cdl.Candle, arg cdl.CandleArg, period int, w float64) *EMA {
	n := len(candles)
	if n == 0 || period <= 0 {
		return nil
	}

	res := make([]float64, n)
	res[0] = candles[0].Arg(arg)
	alpha := w / (float64(period) + w - 1)

	for i := 1; i < n; i++ {
		res[i] = candles[i].Arg(arg)*alpha + res[i-1]*(1-alpha)
	}

	return &EMA{
		baseMA:    baseMA{Res: res, Len: n, Period: period},
		alpha:     alpha,
		CandleArg: arg,
	}
}

type VWMA struct {
	baseMA
	CandleArg   cdl.CandleArg
	sumPriceVol float64
	sumVolume   float64
}

func (v *VWMA) Next(candles []cdl.Candle) {
	n := len(candles)
	idx := n - 1
	oldIdx := n - min(n, v.Period+1)

	price := candles[idx].Arg(v.CandleArg)
	volume := candles[idx].Volume
	oldPrice := candles[oldIdx].Arg(v.CandleArg)
	oldVolume := candles[oldIdx].Volume

	v.sumPriceVol += (price * volume) - (oldPrice * oldVolume)
	v.sumVolume += volume - oldVolume

	v.Res = append(v.Res, v.sumPriceVol/v.sumVolume)
	v.Len++
}

func NewVWMA(candles []cdl.Candle, arg cdl.CandleArg, period int) *VWMA {
	n := len(candles)
	if n == 0 || period <= 0 {
		return nil
	}

	res := make([]float64, n)
	var sumPriceVol float64
	var sumVolume float64

	for i := 0; i < period && i < n; i++ {
		price := candles[i].Arg(arg)
		volume := candles[i].Volume
		sumPriceVol += price * volume
		sumVolume += volume
		res[i] = sumPriceVol / sumVolume
	}

	for i := period; i < n; i++ {
		price := candles[i].Arg(arg)
		volume := candles[i].Volume
		oldPrice := candles[i-period].Arg(arg)
		oldVolume := candles[i-period].Volume

		sumPriceVol += (price * volume) - (oldPrice * oldVolume)
		sumVolume += volume - oldVolume

		res[i] = sumPriceVol / sumVolume
	}

	return &VWMA{
		baseMA:      baseMA{Res: res, Len: n, Period: period},
		CandleArg:   arg,
		sumPriceVol: sumPriceVol,
		sumVolume:   sumVolume,
	}
}


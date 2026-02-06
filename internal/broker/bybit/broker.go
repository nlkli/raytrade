package bybit

import (
	"encoding/json"
	"fmt"
	"nlkli/raytrade/internal/broker"
	"nlkli/raytrade/internal/broker/bybit/models"
	"nlkli/raytrade/internal/cdl"
	"slices"
	"strconv"
)

type Broker struct {
	c *Client

	publicStream map[broker.Category]*Stream
}

func NewBroker(c *Client) *Broker {
	return &Broker{
		c:            c,
		publicStream: make(map[broker.Category]*Stream, 2),
	}
}

func (b *Broker) CandleStream(
	done <-chan struct{},
	category broker.Category,
	symbol string,
	interval cdl.Interval,
) (<-chan cdl.CandleStreamData, error) {
	lc := ToLocalCategory(category)
	li, err := TryToLocalInterval(interval)
	if err != nil {
		return nil, err
	}

	if _, ok := b.publicStream[category]; !ok {
		b.publicStream[category] = b.c.CreatePublicStream(lc)
	}

	topic := fmt.Sprintf("kline.%s.%s", li, symbol)
	sub, err := b.publicStream[category].Subscribe([]string{topic}, 1)
	if err != nil {
		return nil, err
	}

	ch := make(chan cdl.CandleStreamData, 1)
	go func() {
		defer close(ch)
		defer sub.Stop()

		for {
			select {
			case data := <-sub.C():
				if data == nil {
					continue
				}
				var streamKlineData models.StreamKlineData
				json.Unmarshal(data.Data, &streamKlineData)
				var candle cdl.Candle
				for _, i := range streamKlineData {
					candle.Time = i.Start
					candle.O, err = strconv.ParseFloat(i.Open, 64)
					if err != nil {
						continue
					}
					candle.H, err = strconv.ParseFloat(i.High, 64)
					if err != nil {
						continue
					}
					candle.L, err = strconv.ParseFloat(i.Low, 64)
					if err != nil {
						continue
					}
					candle.C, err = strconv.ParseFloat(i.Close, 64)
					if err != nil {
						continue
					}
					candle.Volume, err = strconv.ParseFloat(i.Volume, 64)
					if err != nil {
						continue
					}
					ch <- cdl.CandleStreamData{
						Candle:   candle,
						Interval: interval,
						Confirm:  i.Confirm,
					}

				}
			case <-done:
				return
			case <-b.c.ctx.Done():
				return
			}
		}

	}()

	return ch, nil
}

func (b *Broker) GetCandles(
	category broker.Category,
	symbol string,
	interval cdl.Interval,
	limit int,
	start *int,
	end *int,
) ([]cdl.Candle, error) {
	if limit == 0 {
		return make([]cdl.Candle, 0), nil
	}

	li, err := TryToLocalInterval(interval)
	if err != nil {
		return nil, err
	}

	res, err := b.c.GetKline(ToLocalCategory(category), symbol, li, start, end, &limit)
	if err != nil {
		return nil, err
	}

	candles := make([]cdl.Candle, len(res.List))
	defer slices.Reverse(candles)

	for n, i := range res.List {
		candles[n].Time, err = strconv.ParseInt(i[0], 10, 64)
		if err != nil {
			return candles, err
		}
		candles[n].O, err = strconv.ParseFloat(i[1], 64)
		if err != nil {
			return candles, err
		}
		candles[n].H, err = strconv.ParseFloat(i[2], 64)
		if err != nil {
			return candles, err
		}
		candles[n].L, err = strconv.ParseFloat(i[3], 64)
		if err != nil {
			return candles, err
		}
		candles[n].C, err = strconv.ParseFloat(i[4], 64)
		if err != nil {
			return candles, err
		}
		candles[n].Volume, err = strconv.ParseFloat(i[5], 64)
		if err != nil {
			return candles, err
		}
	}

	return candles, nil
}

func (b *Broker) ExtendStartCandles(
	candles []cdl.Candle,
	category broker.Category,
	symbol string,
	interval cdl.Interval,
	limit int,
) ([]cdl.Candle, error) {
	if limit == 0 {
		return candles, nil
	}

	var end *int
	if len(candles) > 0 {
		t := int(candles[0].Time) - 29999
		end = &t
	}

	start, err := b.GetCandles(category, symbol, interval, limit, nil, end)
	if err != nil {
		return candles, err
	}

	res := make([]cdl.Candle, 0, len(start)+len(candles))
	res = append(res, start...)
	res = append(res, candles...)

	return res, nil
}

func (b *Broker) ExtendEndCandles(
	candles []cdl.Candle,
	category broker.Category,
	symbol string,
	interval cdl.Interval,
	limit int,
) ([]cdl.Candle, error) {
	if limit == 0 {
		return candles, nil
	}

	var start *int
	if len(candles) > 0 {
		t := int(candles[len(candles)-1].Time) + 29999
		start = &t
	}

	end, err := b.GetCandles(category, symbol, interval, limit, start, nil)
	if err != nil {
		return candles, err
	}

	res := make([]cdl.Candle, 0, len(candles)+len(end))
	res = append(res, candles...)
	res = append(res, end...)

	return res, nil
}

func ToLocalCategory(c broker.Category) models.Category {
	switch c {
	case broker.Spot:
		return models.CategorySpot
	case broker.Futures:
		return models.CategoryLinear
	default:
		return models.CategoryDefault
	}
}

func TryToLocalInterval(i cdl.Interval) (models.Interval, error) {
	switch i {
	case cdl.M1:
		return models.Interval1Min, nil
	case cdl.M3:
		return models.Interval3Min, nil
	case cdl.M5:
		return models.Interval5Min, nil
	case cdl.M15:
		return models.Interval15Min, nil
	case cdl.M30:
		return models.Interval30Min, nil
	case cdl.H1:
		return models.Interval60Min, nil
	case cdl.H2:
		return models.Interval120Min, nil
	case cdl.H4:
		return models.Interval240Min, nil
	case cdl.H6:
		return models.Interval360Min, nil
	case cdl.H12:
		return models.Interval720Min, nil
	case cdl.D1:
		return models.Interval1Day, nil
	case cdl.D7:
		return models.Interval1Week, nil
	case cdl.D30:
		return models.Interval1Month, nil
	default:
		return "", fmt.Errorf("unknown cdl interval: %d", i)
	}
}

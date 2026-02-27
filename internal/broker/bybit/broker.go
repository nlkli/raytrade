package bybit

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"nlkli/raytrade/internal/broker"
	"nlkli/raytrade/internal/broker/bybit/models"
	"nlkli/raytrade/internal/cdl"
	"nlkli/raytrade/internal/ws"
	"slices"
	"sort"
	"strconv"
	"sync"
	"time"
)

type Broker struct {
	c *Client
}

func NewBroker(c *Client) *Broker {
	return &Broker{
		c: c,
	}
}

func (b *Broker) GetCandles(

	ctx context.Context,

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

	res, err := b.c.GetKline(
		ctx, ToLocalCategory(category), symbol, li, start, end, &limit,
	)
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

	ctx context.Context,

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

	start, err := b.GetCandles(
		ctx, category, symbol, interval, limit, nil, end,
	)
	if err != nil {
		return candles, err
	}

	res := make([]cdl.Candle, 0, len(start)+len(candles))
	res = append(res, start...)
	res = append(res, candles...)

	return res, nil
}

func (b *Broker) ExtendEndCandles(

	ctx context.Context,

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

	end, err := b.GetCandles(
		ctx, category, symbol, interval, limit, start, nil,
	)
	if err != nil {
		return candles, err
	}

	res := make([]cdl.Candle, 0, len(candles)+len(end))
	res = append(res, candles...)
	res = append(res, end...)

	return res, nil
}

func (b *Broker) PlaceOrder(

	ctx context.Context,
	category broker.Category,
	symbol string,
	side broker.Side,
	price float64,
	usdQty float64,

) {

}

func (b *Broker) GetOpenOrders(

	ctx context.Context,
	category broker.Category,
	symbol string,
	limit int,

) ([]broker.Order, string, error) {
	res, err := b.c.GetOrderList(
		ctx,
		ToLocalCategory(category),
		&GetOrderListRequestParas{
			Symbol: &symbol,
			Limit:  &limit,
			// OpenOnly is default
		},
	)
	if err != nil {
		return nil, "", err
	}

	orderList := make([]broker.Order, len(res.List))

	if len(res.List) == 0 {
		return orderList, "", nil
	}

	for i, o := range res.List {
		order := &orderList[i]

		order.Symbol = o.Symbol

		switch o.Side {
		case "Buy":
			order.Side = broker.Long
		case "Sell":
			order.Side = broker.Short
		}

		switch o.OrderStatus {
		case models.OrderStatusNew,
			models.OrderStatusPartiallyFilled,
			models.OrderStatusUntriggered:
			order.Status = broker.Open
		default:
			order.Status = broker.Closed
		}

		order.Price, err = strconv.ParseFloat(o.Price, 64)
		if err != nil {
			return nil, "", err
		}

		order.Qty, err = strconv.ParseFloat(o.Qty, 64)
		if err != nil {
			return nil, "", err
		}

		order.ExecQty, err = strconv.ParseFloat(o.CumExecQty, 64)
		if err != nil {
			return nil, "", err
		}

		order.ExecValue, err = strconv.ParseFloat(o.CumExecValue, 64)
		if err != nil {
			return nil, "", err
		}

		if len(o.AvgPrice) != 0 {
			ep, err := strconv.ParseFloat(o.AvgPrice, 64)
			if err != nil {
				return nil, "", err
			}
			order.EntryPrice = &ep
		}

		createdAtUnix, err := strconv.ParseInt(o.CreatedTime, 0, 64)
		if err != nil {
			return nil, "", err
		}

		order.CreatedAt = time.UnixMilli(createdAtUnix)
	}

	return orderList, res.NextPageCursor, nil
}

func (b *Broker) GetPosition(

	ctx context.Context,
	category broker.Category,
	symbol string,

) (*broker.Position, error) {

	res, err := b.c.GetPositionInfo(
		ctx,
		ToLocalCategory(category),
		&symbol,
		nil,
		nil,
	)
	if err != nil {
		return nil, err
	}

	if len(res.List) == 0 {
		return nil, errors.New("position not found")
	}

	pi := res.List[0]

	var side broker.Side
	switch pi.Side {
	case "Buy":
		side = broker.Long
	case "Sell":
		side = broker.Short
	}

	size, err := strconv.ParseFloat(pi.Size, 64)
	if err != nil {
		return nil, err
	}

	entryPrice, err := strconv.ParseFloat(pi.AvgPrice, 64)
	if err != nil {
		return nil, err
	}

	createdAtUnix, err := strconv.ParseInt(pi.CreatedTime, 0, 64)
	if err != nil {
		return nil, err
	}
	createdAt := time.UnixMilli(createdAtUnix)

	return &broker.Position{
		Symbol:     pi.Symbol,
		Side:       side,
		Size:       size,
		EntryPrice: entryPrice,
		CreatedAt:  createdAt,
	}, nil
}

// [bids, asks][][price, size]
func (b *Broker) GetOrderBook(

	ctx context.Context,
	category broker.Category,
	symbol string,
	limit int,

) (*[2][][2]float64, error) {

	res, err := b.c.GetOrderBook(
		ctx, ToLocalCategory(category), symbol, &limit,
	)
	if err != nil {
		return nil, err
	}

	n := min(len(res.B), len(res.A))

	bids := make([][2]float64, n)
	asks := make([][2]float64, n)

	ob := [2][][2]float64{bids, asks}

	for i := range n {
		var bid [2]float64

		bid[0], err = strconv.ParseFloat(res.B[i][0], 64)
		if err != nil {
			return &ob, err
		}

		bid[1], err = strconv.ParseFloat(res.B[i][1], 64)
		if err != nil {
			return &ob, err
		}

		var ask [2]float64

		ask[0], err = strconv.ParseFloat(res.A[i][0], 64)
		if err != nil {
			return &ob, err
		}

		ask[1], err = strconv.ParseFloat(res.A[i][1], 64)
		if err != nil {
			return &ob, err
		}

		ob[0][i] = bid
		ob[1][i] = ask
	}

	return &ob, nil
}

func (b *Broker) CreateStream(

	category broker.Category,
	onConnected ws.OnConnectedFn,

	opts ...ws.PolicyOption,

) broker.Stream {

	lc := ToLocalCategory(category)
	tx := make(chan []byte, 1)
	stream := b.c.CreatePublicStreamV2(lc, tx, onConnected, opts...)

	return &BrokerStream{
		stream: stream,
		tx:     tx,
	}
}

func (b *Broker) CreatePrivateStream(

	onConnected ws.OnConnectedFn,
	opts ...ws.PolicyOption,

) broker.PrivateStream {

	tx := make(chan []byte, 1)
	stream := b.c.CreatePrivateStreamV2(tx, onConnected, opts...)

	return &BrokerPrivateStream{
		stream: stream,
		tx:     tx,
	}
}

type BrokerPrivateStream struct {
	tx     chan []byte
	once   sync.Once
	stream *StreamV2
}

func (s *BrokerPrivateStream) Close() {
	s.once.Do(func() { close(s.tx) })
}

func (s *BrokerPrivateStream) SubscribePosition() (*broker.Subscription[broker.Position], error) {
	subCh, err := s.stream.Subscribe("position") // All-In-One Topic
	if err != nil {
		return nil, err
	}

	ch := make(chan broker.Position, 1)
	go func() {
		defer close(ch)

		for d := range subCh {

			var pi models.StreamPositionInfo
			if err := json.Unmarshal(d.Data, &pi); err != nil {
				continue
			}

			var pos broker.Position

			switch pi.Side {
			case "Buy":
				pos.Side = broker.Long
			case "Sell":
				pos.Side = broker.Short
			}

			pos.Size, err = strconv.ParseFloat(pi.Size, 64)
			if err != nil {
				continue
			}

			pos.EntryPrice, err = strconv.ParseFloat(pi.EntryPrice, 64)
			if err != nil {
				continue
			}

			createdAtUnix, err := strconv.ParseInt(pi.CreatedTime, 0, 64)
			if err != nil {
				continue
			}
			pos.CreatedAt = time.UnixMilli(createdAtUnix)

			ch <- pos
		}
	}()
	return nil, nil
}

type BrokerStream struct {
	tx     chan []byte
	once   sync.Once
	stream *StreamV2
}

func (s *BrokerStream) Close() {
	s.once.Do(func() { close(s.tx) })
}

func (s *BrokerStream) SubscribeCandle(

	symbol string,
	interval cdl.Interval,

) (*broker.Subscription[cdl.CandleStreamData], error) {

	li, err := TryToLocalInterval(interval)
	if err != nil {
		return nil, err
	}

	topic := fmt.Sprintf("kline.%s.%s", li, symbol)
	subCh, err := s.stream.Subscribe(topic)
	if err != nil {
		return nil, err
	}

	ch := make(chan cdl.CandleStreamData, 1)
	go func() {
		defer close(ch)

		for d := range subCh {

			var streamKlineData models.StreamKlineData
			json.Unmarshal(d.Data, &streamKlineData)

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
		}
	}()

	return broker.NewBrokerStreamSubscription(
		ch,
		func() error {
			return s.stream.Unsubscribe(topic)
		},
	), nil
}

// https://bybit-exchange.github.io/docs/v5/websocket/public/orderbook
func (s *BrokerStream) SubscribeOrderBook(

	symbol string, depth int,

) (*broker.Subscription[[2][][2]float64], error) {

	switch depth {
	case 1, 50, 200, 1000:
	default:
		return nil, fmt.Errorf("invalid depth: %d, must be one of: 1, 50, 200, 1000", depth)
	}

	topic := fmt.Sprintf("orderbook.%d.%s", depth, symbol)
	subCh, err := s.stream.Subscribe(topic)
	if err != nil {
		return nil, err
	}

	ch := make(chan [2][][2]float64, 1)
	go func() {
		defer close(ch)

		// [bids, asks][][price, size]
		ob := [2][][2]float64{
			make([][2]float64, 0, 100),
			make([][2]float64, 0, 100),
		}

		for d := range subCh {

			var obData models.StreamOrderBookData
			if err := json.Unmarshal(d.Data, &obData); err != nil {
				continue
			}

			for len(ob[0]) > 0 && len(ob[1]) > 0 && ob[0][0][0] >= ob[1][0][0] {
				ob[0] = ob[0][1:]
				ob[1] = ob[1][1:]
			}

			switch d.Type {
			case "snapshot":
				ob[0] = ob[0][:0]
				ob[1] = ob[1][:0]

				for _, bid := range obData.Bids {
					price, _ := strconv.ParseFloat(bid[0], 64)
					size, _ := strconv.ParseFloat(bid[1], 64)
					ob[0] = append(ob[0], [2]float64{price, size})
				}

				for _, ask := range obData.Asks {
					price, _ := strconv.ParseFloat(ask[0], 64)
					size, _ := strconv.ParseFloat(ask[1], 64)
					ob[1] = append(ob[1], [2]float64{price, size})
				}

			case "delta":
				if bids := obData.Bids; len(bids) > 0 {
					for _, bid := range bids {
						price, _ := strconv.ParseFloat(bid[0], 64)
						size, _ := strconv.ParseFloat(bid[1], 64)

						n := len(ob[0])

						i := sort.Search(n, func(j int) bool {
							return ob[0][j][0] <= price
						})

						if i < n && ob[0][i][0] == price {
							if size == 0 {
								copy(ob[0][i:], ob[0][i+1:n])
								ob[0] = ob[0][:n-1]
							} else {
								ob[0][i][1] = size
							}
						} else if size != 0 {
							ob[0] = append(ob[0], [2]float64{})
							copy(ob[0][i+1:], ob[0][i:n])
							ob[0][i] = [2]float64{price, size}
						}
					}

					if len(ob[0]) > depth {
						ob[0] = ob[0][:depth]
					}

				}

				if asks := obData.Asks; len(asks) > 0 {
					for _, ask := range asks {
						price, _ := strconv.ParseFloat(ask[0], 64)
						size, _ := strconv.ParseFloat(ask[1], 64)

						n := len(ob[1])

						i := sort.Search(n, func(j int) bool {
							return ob[1][j][0] >= price
						})

						if i < n && ob[1][i][0] == price {
							if size == 0 {
								copy(ob[1][i:], ob[1][i+1:n])
								ob[1] = ob[1][:n-1]
							} else {
								ob[1][i][1] = size
							}
						} else if size != 0 {
							ob[1] = append(ob[1], [2]float64{})
							copy(ob[1][i+1:], ob[1][i:n])
							ob[1][i] = [2]float64{price, size}
						}
					}

					if len(ob[1]) > depth {
						ob[1] = ob[1][:depth]
					}
				}
			}

			bidsCopy := make([][2]float64, len(ob[0]))
			copy(bidsCopy, ob[0])

			asksCopy := make([][2]float64, len(ob[1]))
			copy(asksCopy, ob[1])

			ch <- [2][][2]float64{bidsCopy, asksCopy}
		}
	}()

	return broker.NewBrokerStreamSubscription(
		ch,
		func() error {
			return s.stream.Unsubscribe(topic)
		},
	), nil
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

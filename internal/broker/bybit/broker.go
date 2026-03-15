package bybit

import (
	"context"
	"encoding/json"
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

// id linkId
func (b *Broker) PlaceOrder(

	ctx context.Context,
	LinkId string,
	params broker.PlaceOrderParams,

) (string, string, error) {

	rp := &PlaceOrderRequestParams{
		Category: ToLocalCategory(params.Category),
		Symbol:   params.Symbol,
	}

	if params.Side == broker.Long {
		rp.Side = models.SideBuy
	} else {
		rp.Side = models.SideSell
	}

	if params.IsLeverage != nil {
		var il int
		if *params.IsLeverage {
			il = 1
		} else {
			il = 0
		}
		rp.IsLeverage = &il
	}

	switch params.Type {
	case broker.Limit:
		rp.OrderType = models.OrderTypeLimit
	case broker.Market:
		rp.OrderType = models.OrderTypeMarket
	}

	rp.Qty = strconv.FormatFloat(params.Qty, 'f', -1, 64)

	if params.MarketUnit != nil {
		var mu models.MarketUnit
		switch *params.MarketUnit {
		case broker.BaseCoin:
			mu = models.MarketUnitBaseCoin
		case broker.QuoteCoin:
			mu = models.MarketUnitQuoteCoin
		}
		rp.MarketUnit = &mu
	}

	if params.Price != nil {
		price := strconv.FormatFloat(*params.Price, 'f', -1, 64)
		rp.Price = &price
	}

	if params.TriggerDirection != nil {
		var td models.TriggerDirection
		switch *params.TriggerDirection {
		case broker.Rise:
			td = models.TriggerDirectionRise
		case broker.Fall:
			td = models.TriggerDirectionFall
		}
		rp.TriggerDirection = &td
	}

	if params.TriggerPrice != nil {
		triggerPrice := strconv.FormatFloat(*params.TriggerPrice, 'f', -1, 64)
		rp.TriggerPrice = &triggerPrice
	}

	if params.TriggerBy != nil {
		var tb models.TriggerBy
		switch *params.TriggerBy {
		case broker.IndexPrice:
			tb = models.TriggerByIndexPrice
		case broker.MarkPrice:
			tb = models.TriggerByMarkPrice
		case broker.LastPrice:
			tb = models.TriggerByLastPrice
		}
		rp.TriggerBy = &tb
	}

	if params.TakeProfit != nil {
		tp := strconv.FormatFloat(*params.TakeProfit, 'f', -1, 64)
		rp.TakeProfit = &tp
	}

	if params.StopLoss != nil {
		sl := strconv.FormatFloat(*params.StopLoss, 'f', -1, 64)
		rp.StopLoss = &sl
	}

	if params.TpTriggerBy != nil {
		var tb models.TriggerBy
		switch *params.TpTriggerBy {
		case broker.IndexPrice:
			tb = models.TriggerByIndexPrice
		case broker.MarkPrice:
			tb = models.TriggerByMarkPrice
		case broker.LastPrice:
			tb = models.TriggerByLastPrice
		}
		rp.TpTriggerBy = &tb
	}

	if params.SlTriggerBy != nil {
		var tb models.TriggerBy
		switch *params.SlTriggerBy {
		case broker.IndexPrice:
			tb = models.TriggerByIndexPrice
		case broker.MarkPrice:
			tb = models.TriggerByMarkPrice
		case broker.LastPrice:
			tb = models.TriggerByLastPrice
		}
		rp.SlTriggerBy = &tb
	}

	if params.ReduceOnly != nil {
		ro := *params.ReduceOnly
		rp.ReduceOnly = &ro
	}

	if params.CloseOnTrigger != nil {
		cot := *params.CloseOnTrigger
		rp.CloseOnTrigger = &cot
	}

	if params.TpslMode != nil {
		var m models.TpslMode
		switch *params.TpslMode {
		case broker.Full:
			m = models.TpslModeFull
		case broker.Partial:
			m = models.TpslModePartial
		}
		rp.TpslMode = &m
	}

	if params.TpLimitPrice != nil {
		tplp := strconv.FormatFloat(*params.TpLimitPrice, 'f', -1, 64)
		rp.TpLimitPrice = &tplp
	}

	if params.SlLimitPrice != nil {
		sllp := strconv.FormatFloat(*params.SlLimitPrice, 'f', -1, 64)
		rp.SlLimitPrice = &sllp
	}

	if params.TpOrderType != nil {
		var ot models.OrderType
		switch *params.TpOrderType {
		case broker.Limit:
			ot = models.OrderTypeLimit
		case broker.Market:
			ot = models.OrderTypeMarket
		}
		rp.TpOrderType = &ot
	}

	if params.SlOrderType != nil {
		var ot models.OrderType
		switch *params.SlOrderType {
		case broker.Limit:
			ot = models.OrderTypeLimit
		case broker.Market:
			ot = models.OrderTypeMarket
		}
		rp.SlOrderType = &ot
	}

	res, err := b.c.PlaceOrder(ctx, rp)
	if err != nil {
		return "", "", err
	}

	return res.OrderId, res.OrderLinkId, nil
}

func (b *Broker) CancelOrder(

	ctx context.Context,
	category broker.Category,
	symbol string,
	id string,
	linkId string,

) (string, string, error) {
	params := &CancelOrderRequestParams{
		Category: ToLocalCategory(category),
		Symbol:   symbol,
	}

	if len(id) > 0 {
		params.OrderId = &id
	}

	if len(linkId) > 0 {
		params.OrderLinkId = &linkId
	}

	res, err := b.c.CancelOrder(ctx, params)
	if err != nil {
		return "", "", err
	}

	return res.OrderId, res.OrderLinkId, nil
}

func (b *Broker) GetOpenOrder(

	ctx context.Context,
	category broker.Category,
	symbol string,

) ([]broker.Order, string, error) {

	limit := MAX_ORDERLIST_LIMIT

	res, err := b.c.GetOrderList(
		ctx,
		&GetOrderListRequestParams{
			Category: ToLocalCategory(category),
			Symbol:   &symbol,
			Limit:    &limit,
			// OpenOnly is default
		},
	)
	if err != nil {
		return nil, "", err
	}

	order := make([]broker.Order, len(res.List))

	for i, oi := range res.List {
		o := &order[i]

		o.Category = FromLocalCategory(res.Category)
		o.Symbol = oi.Symbol

		o.Id = oi.OrderId
		o.LinkId = oi.OrderLinkId
		o.Price, _ = strconv.ParseFloat(oi.Price, 64)
		o.Qty, _ = strconv.ParseFloat(oi.Qty, 64)

		if oi.MarketUnit == models.MarketUnitBaseCoin {
			o.MarketUnit = broker.BaseCoin
		} else {
			o.MarketUnit = broker.QuoteCoin
		}

		if oi.Side == models.SideBuy {
			o.Side = broker.Long
		} else {
			o.Side = broker.Short
		}

		o.IsLeverage = oi.IsLeverage == "1"

		switch oi.OrderStatus {
		case models.OrderStatusNew,
			models.OrderStatusPartiallyFilled,
			models.OrderStatusUntriggered:
			o.Status = broker.Open
		default:
			o.Status = broker.Closed
		}

		if oi.OrderType == models.OrderTypeLimit {
			o.Type = broker.Limit
		} else {
			o.Type = broker.Market
		}

		o.AvgPrice, _ = strconv.ParseFloat(oi.AvgPrice, 64)
		o.LeavesQty, _ = strconv.ParseFloat(oi.LeavesQty, 64)
		o.LeavesValue, _ = strconv.ParseFloat(oi.LeavesValue, 64)
		o.ExecQty, _ = strconv.ParseFloat(oi.CumExecQty, 64)
		o.ExecValue, _ = strconv.ParseFloat(oi.CumExecValue, 64)
		o.TriggerPrice, _ = strconv.ParseFloat(oi.TriggerPrice, 64)
		o.TakeProfit, _ = strconv.ParseFloat(oi.TakeProfit, 64)
		o.StopLoss, _ = strconv.ParseFloat(oi.StopLoss, 64)

		switch oi.TpslMode {
		case models.TpslModeFull:
			o.TpslMode = broker.Full
		case models.TpslModePartial:
			o.TpslMode = broker.Partial
		default:
			o.TpslMode = -1
		}

		o.TpLimitPrice, _ = strconv.ParseFloat(oi.TpLimitPrice, 64)
		o.SlLimitPrice, _ = strconv.ParseFloat(oi.SlLimitPrice, 64)

		switch oi.TpTriggerBy {
		case models.TriggerByIndexPrice:
			o.TpTriggerBy = broker.IndexPrice
		case models.TriggerByMarkPrice:
			o.TpTriggerBy = broker.MarkPrice
		case models.TriggerByLastPrice:
			o.TpTriggerBy = broker.LastPrice
		default:
			o.TpTriggerBy = -1
		}

		switch oi.SlTriggerBy {
		case models.TriggerByIndexPrice:
			o.SlTriggerBy = broker.IndexPrice
		case models.TriggerByMarkPrice:
			o.SlTriggerBy = broker.MarkPrice
		case models.TriggerByLastPrice:
			o.SlTriggerBy = broker.LastPrice
		default:
			o.SlTriggerBy = -1
		}

		switch oi.TriggerDirection {
		case models.TriggerDirectionFall:
			o.TriggerDirection = broker.Fall
		case models.TriggerDirectionRise:
			o.TriggerDirection = broker.Rise
		default:
			o.TriggerDirection = -1
		}

		switch oi.TriggerBy {
		case models.TriggerByIndexPrice:
			o.TriggerBy = broker.IndexPrice
		case models.TriggerByMarkPrice:
			o.TriggerBy = broker.MarkPrice
		case models.TriggerByLastPrice:
			o.TriggerBy = broker.LastPrice
		default:
			o.TriggerBy = -1
		}

		o.ReduceOnly = oi.ReduceOnly
		o.CloseOnTrigger = oi.CloseOnTrigger

		createdAtUnix, _ := strconv.ParseInt(oi.CreatedTime, 0, 64)
		o.CreatedAt = time.UnixMilli(createdAtUnix)
	}

	return order, res.NextPageCursor, nil
}

func (b *Broker) GetPosition(

	ctx context.Context,
	category broker.Category,
	symbol string,

) ([]broker.Position, error) {

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

	position := make([]broker.Position, len(res.List))

	for i, pi := range res.List {
		p := &position[i]

		p.Category = FromLocalCategory(res.Category)
		p.Symbol = pi.Symbol
		if pi.Side == models.SideBuy {
			p.Side = broker.Long
		} else {
			p.Side = broker.Short
		}
		p.Size, _ = strconv.ParseFloat(pi.Size, 64)
		p.EntryPrice, _ = strconv.ParseFloat(pi.AvgPrice, 64)
		p.PositionValue, _ = strconv.ParseFloat(pi.PositionValue, 64)
		p.PositionIM, _ = strconv.ParseFloat(pi.PositionIM, 64)
		p.Leverage, _ = strconv.Atoi(pi.Leverage)
		p.MarkPrice, _ = strconv.ParseFloat(pi.MarkPrice, 64)
		p.BreakEvenPrice, _ = strconv.ParseFloat(pi.BreakEvenPrice, 64)
		p.UnrealisedPnl, _ = strconv.ParseFloat(pi.UnrealisedPnl, 64)
		p.RealisedPnl, _ = strconv.ParseFloat(pi.CurRealisedPnl, 64)
		p.LiqPrice, _ = strconv.ParseFloat(pi.LiqPrice, 64)
		p.TakeProfit, _ = strconv.ParseFloat(pi.TakeProfit, 64)
		p.StopLoss, _ = strconv.ParseFloat(pi.StopLoss, 64)

		createdAtUnix, _ := strconv.ParseInt(pi.CreatedTime, 0, 64)
		p.CreatedAt = time.UnixMilli(createdAtUnix)
	}

	return position, nil
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

type BrokerPrivateStream struct {
	tx     chan []byte
	once   sync.Once
	stream *StreamV2
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

func (s *BrokerPrivateStream) Close() {
	s.once.Do(func() { close(s.tx) })
}

func (s *BrokerPrivateStream) SubscribeExecution() (*broker.Subscription[broker.Execution], error) {
	subCh, err := s.stream.Subscribe("execution") // All-In-One Topic
	if err != nil {
		return nil, err
	}

	ch := make(chan broker.Execution, 1)
	go func() {
		defer close(ch)

		for d := range subCh {

			var exList []models.StreamExecutionInfo
			if err := json.Unmarshal(d.Data, &exList); err != nil {
				continue
			}

			for _, ei := range exList {
				var e broker.Execution

				e.Category = FromLocalCategory(ei.Category)
				e.Symbol = ei.Symbol
				if ei.Side == models.SideBuy {
					e.Side = broker.Long
				} else {
					e.Side = broker.Short
				}
				e.Qty, _ = strconv.ParseFloat(ei.ExecQty, 64)
				e.Price, _ = strconv.ParseFloat(ei.ExecPrice, 64)
				e.OrderId = ei.OrderId
				e.OrderLinkId = ei.OrderLinkId
				e.IsMaker = ei.IsMaker

				createdAtUnix, _ := strconv.ParseInt(ei.ExecTime, 0, 64)
				e.Time = time.UnixMilli(createdAtUnix)

				ch <- e
			}
		}
	}()

	return broker.NewBrokerStreamSubscription(
		ch,
		func() error {
			return s.stream.Unsubscribe("execution")
		},
	), nil
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

			var piList []models.StreamPositionInfo
			if err := json.Unmarshal(d.Data, &piList); err != nil {
				continue
			}

			for _, pi := range piList {
				var p broker.Position

				p.Category = FromLocalCategory(pi.Category)
				p.Symbol = pi.Symbol
				if pi.Side == models.SideBuy {
					p.Side = broker.Long
				} else {
					p.Side = broker.Short
				}
				p.Size, _ = strconv.ParseFloat(pi.Size, 64)
				p.EntryPrice, _ = strconv.ParseFloat(pi.EntryPrice, 64)
				p.PositionValue, _ = strconv.ParseFloat(pi.PositionValue, 64)
				p.PositionIM, _ = strconv.ParseFloat(pi.PositionIM, 64)
				p.Leverage, _ = strconv.Atoi(pi.Leverage)
				p.MarkPrice, _ = strconv.ParseFloat(pi.MarkPrice, 64)
				p.BreakEvenPrice, _ = strconv.ParseFloat(pi.BreakEvenPrice, 64)
				p.UnrealisedPnl, _ = strconv.ParseFloat(pi.UnrealisedPnl, 64)
				p.RealisedPnl, _ = strconv.ParseFloat(pi.CurRealisedPnl, 64)
				p.LiqPrice, _ = strconv.ParseFloat(pi.LiqPrice, 64)
				p.TakeProfit, _ = strconv.ParseFloat(pi.TakeProfit, 64)
				p.StopLoss, _ = strconv.ParseFloat(pi.StopLoss, 64)

				createdAtUnix, _ := strconv.ParseInt(pi.CreatedTime, 0, 64)
				p.CreatedAt = time.UnixMilli(createdAtUnix)

				ch <- p
			}
		}
	}()

	return broker.NewBrokerStreamSubscription(
		ch,
		func() error {
			return s.stream.Unsubscribe("position")
		},
	), nil
}

func (s *BrokerPrivateStream) SubscribeOrder() (*broker.Subscription[broker.Order], error) {
	subCh, err := s.stream.Subscribe("order") // All-In-One Topic
	if err != nil {
		return nil, err
	}

	ch := make(chan broker.Order, 1)
	go func() {
		defer close(ch)

		for d := range subCh {

			var oiList []models.StreamOrderInfo
			if err := json.Unmarshal(d.Data, &oiList); err != nil {
				continue
			}

			for _, oi := range oiList {
				var o broker.Order

				o.Category = FromLocalCategory(oi.Category)
				o.Symbol = oi.Symbol

				o.Id = oi.OrderId
				o.LinkId = oi.OrderLinkId
				o.Price, _ = strconv.ParseFloat(oi.Price, 64)
				o.Qty, _ = strconv.ParseFloat(oi.Qty, 64)

				if oi.MarketUnit == models.MarketUnitBaseCoin {
					o.MarketUnit = broker.BaseCoin
				} else {
					o.MarketUnit = broker.QuoteCoin
				}

				if oi.Side == models.SideBuy {
					o.Side = broker.Long
				} else {
					o.Side = broker.Short
				}

				o.IsLeverage = oi.IsLeverage == "1"

				switch oi.OrderStatus {
				case models.OrderStatusNew,
					models.OrderStatusPartiallyFilled,
					models.OrderStatusUntriggered:
					o.Status = broker.Open
				default:
					o.Status = broker.Closed
				}

				if oi.OrderType == models.OrderTypeLimit {
					o.Type = broker.Limit
				} else {
					o.Type = broker.Market
				}

				o.AvgPrice, _ = strconv.ParseFloat(oi.AvgPrice, 64)
				o.LeavesQty, _ = strconv.ParseFloat(oi.LeavesQty, 64)
				o.LeavesValue, _ = strconv.ParseFloat(oi.LeavesValue, 64)
				o.ExecQty, _ = strconv.ParseFloat(oi.CumExecQty, 64)
				o.ExecValue, _ = strconv.ParseFloat(oi.CumExecValue, 64)
				o.TriggerPrice, _ = strconv.ParseFloat(oi.TriggerPrice, 64)
				o.TakeProfit, _ = strconv.ParseFloat(oi.TakeProfit, 64)
				o.StopLoss, _ = strconv.ParseFloat(oi.StopLoss, 64)

				switch oi.TpslMode {
				case models.TpslModeFull:
					o.TpslMode = broker.Full
				case models.TpslModePartial:
					o.TpslMode = broker.Partial
				default:
					o.TpslMode = -1
				}

				o.TpLimitPrice, _ = strconv.ParseFloat(oi.TpLimitPrice, 64)
				o.SlLimitPrice, _ = strconv.ParseFloat(oi.SlLimitPrice, 64)

				switch oi.TpTriggerBy {
				case models.TriggerByIndexPrice:
					o.TpTriggerBy = broker.IndexPrice
				case models.TriggerByMarkPrice:
					o.TpTriggerBy = broker.MarkPrice
				case models.TriggerByLastPrice:
					o.TpTriggerBy = broker.LastPrice
				default:
					o.TpTriggerBy = -1
				}

				switch oi.SlTriggerBy {
				case models.TriggerByIndexPrice:
					o.SlTriggerBy = broker.IndexPrice
				case models.TriggerByMarkPrice:
					o.SlTriggerBy = broker.MarkPrice
				case models.TriggerByLastPrice:
					o.SlTriggerBy = broker.LastPrice
				default:
					o.SlTriggerBy = -1
				}

				switch oi.TriggerDirection {
				case models.TriggerDirectionFall:
					o.TriggerDirection = broker.Fall
				case models.TriggerDirectionRise:
					o.TriggerDirection = broker.Rise
				default:
					o.TriggerDirection = -1
				}

				switch oi.TriggerBy {
				case models.TriggerByIndexPrice:
					o.TriggerBy = broker.IndexPrice
				case models.TriggerByMarkPrice:
					o.TriggerBy = broker.MarkPrice
				case models.TriggerByLastPrice:
					o.TriggerBy = broker.LastPrice
				default:
					o.TriggerBy = -1
				}

				o.ReduceOnly = oi.ReduceOnly
				o.CloseOnTrigger = oi.CloseOnTrigger

				createdAtUnix, _ := strconv.ParseInt(oi.CreatedTime, 0, 64)
				o.CreatedAt = time.UnixMilli(createdAtUnix)

				ch <- o
			}
		}
	}()

	return broker.NewBrokerStreamSubscription(
		ch,
		func() error {
			return s.stream.Unsubscribe("position")
		},
	), nil
}

type BrokerStream struct {
	tx     chan []byte
	once   sync.Once
	stream *StreamV2
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
			make([][2]float64, 0, depth),
			make([][2]float64, 0, depth),
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

func FromLocalCategory(c models.Category) broker.Category {
	switch c {
	case models.CategoryLinear:
		return broker.Futures
	case models.CategorySpot:
		return broker.Spot
	default:
		return broker.Futures
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

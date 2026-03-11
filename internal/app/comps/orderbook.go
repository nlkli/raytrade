package comps

import (
	"nlkli/raytrade/internal/app/core"
	"strconv"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// TODO
// del showText
// del OrderBookState
// move state to glob

const (
	ORDER_BOOK_TEXT_XPD float32 = 4 // Padding
)

type orderBookLocalOrder struct {
	price float64
}

type OrderBook struct {
	*Rect

	StateIdx int

	rh float32

	ySep float32
	xSep float32

	cap int

	maxBidS,
	maxAskS float64

	bidsText [][2]string
	asksText [][2]string

	bidsPriceTextW []float32
	asksPriceTextW []float32

	localOrder []orderBookLocalOrder
}

func (ob *OrderBook) R() *Rect {
	return ob.Rect
}

func (ob *OrderBook) Render(s *core.State) {
	obS := s.OrderBook[ob.StateIdx]

	switch obS.VM {
	case 0:
		ob.RenderCenteredView(s, obS)
	default:
		ob.RenderSplitView(s, obS)
	}

	if obS.Forced {
		obS.Forced = false
	}
}

func (ob *OrderBook) RenderCenteredView(s *core.State, obS *core.OrderBookState) {
	n := min(len(obS.Bids), len(obS.Asks))
	if n == 0 {
		return
	}

	ob.rh = s.RH - obS.RHD

	halfY := ob.s.Y * .5
	ob.ySep = ob.p.Y + halfY + obS.ShiftY

	if s.WRF || obS.Forced {
		shiftY := obS.ShiftY
		if shiftY < 0 {
			shiftY = -shiftY
		}
		ob.cap = min(n, int((halfY+shiftY)/ob.rh)+1)
		ob.UpdateOrderBookState(s, obS)
	}

	rl.BeginScissorMode(
		int32(ob.p.X),
		int32(ob.p.Y),
		int32(ob.s.X),
		int32(ob.s.Y),
	)

	rl.DrawLineEx(
		rl.Vector2{X: ob.p.X, Y: ob.ySep},
		rl.Vector2{X: ob.p.X + ob.s.X, Y: ob.ySep},
		2,
		s.P.Base.Orange,
	)

	maxSizeW := ob.s.X - 1

	for i := range ob.cap {
		offsetY := 2 + float32(i)*ob.rh

		bid := obS.Bids[i]

		bidSizeRatio := float32(bid[1] / ob.maxBidS)
		bidSizeW := maxSizeW * bidSizeRatio

		bidPos := rl.Vector2{X: ob.p.X, Y: ob.ySep + offsetY}

		rl.DrawRectangleV(
			bidPos,
			rl.Vector2{X: bidSizeW, Y: ob.rh},
			s.P.Diff.Add,
		)

		rl.DrawTextEx(
			s.F,
			ob.bidsText[i][1],
			rl.Vector2{
				X: bidPos.X + ORDER_BOOK_TEXT_XPD,
				Y: bidPos.Y,
			},
			ob.rh,
			0,
			s.P.Fg[2],
		)

		rl.DrawTextEx(
			s.F,
			ob.bidsText[i][0],
			rl.Vector2{
				X: ob.p.X + ob.s.X -
					ob.bidsPriceTextW[i] - ORDER_BOOK_TEXT_XPD,
				Y: bidPos.Y,
			},
			ob.rh,
			0,
			s.P.Fg[1],
		)

		rl.DrawLineEx(
			rl.Vector2{X: ob.p.X, Y: bidPos.Y + ob.rh},
			rl.Vector2{X: ob.p.X + ob.s.X, Y: bidPos.Y + ob.rh},
			1,
			s.P.Bg[0],
		)

		ask := obS.Asks[i]

		askSizeRatio := float32(ask[1] / ob.maxAskS)
		askSizeW := ob.s.X * askSizeRatio

		askPos := rl.Vector2{X: ob.p.X, Y: ob.ySep - offsetY - ob.rh}

		rl.DrawRectangleV(
			askPos,
			rl.Vector2{X: askSizeW, Y: ob.rh},
			s.P.Diff.Delete,
		)

		rl.DrawTextEx(
			s.F,
			ob.asksText[i][1],
			rl.Vector2{
				X: askPos.X + ORDER_BOOK_TEXT_XPD,
				Y: askPos.Y,
			},
			ob.rh,
			0,
			s.P.Fg[2],
		)

		rl.DrawTextEx(
			s.F,
			ob.asksText[i][0],
			rl.Vector2{
				X: ob.p.X + ob.s.X -
					ob.asksPriceTextW[i] - ORDER_BOOK_TEXT_XPD,
				Y: askPos.Y,
			},
			ob.rh,
			0,
			s.P.Fg[1],
		)

		rl.DrawLineEx(
			rl.Vector2{X: ob.p.X, Y: askPos.Y},
			rl.Vector2{X: ob.p.X + ob.s.X, Y: askPos.Y},
			1,
			s.P.Bg[0],
		)
	}
	// 	if obS.ShowOrder {
	//
	// 		if s.Order.Forced {
	// 			ob.localOrder = ob.localOrder[:0]
	// 			for i, oi := range s.Order.List {
	// 				if oi.Category != obS.Category ||
	// 					oi.Symbol != obS.Symbol ||
	// 					oi.Price <= 0 {
	// 					continue
	// 				}
	//
	// 				ob.localOrder = append(
	// 					ob.localOrder,
	// 					orderBookLocalOrder{
	// 						price: oi.Price,
	// 					},
	// 				)
	// 			}
	//
	// 		}
	//
	// 	}
	//
	if !s.E.Mouse.Captured && ob.ContainsV(s.E.Mouse.Pos) {

		if s.E.Mouse.HoldLeft {
			obS.ShiftY -= s.E.Mouse.Delta.Y
		} else {
			if s.E.Mouse.Click[0] {
				if s.E.Mouse.Pos.Y < ob.ySep {
					askIdx := int((ob.ySep - s.E.Mouse.Pos.Y - 1) / ob.rh)
					if len(obS.Asks) > askIdx {
						s.BTX <- &core.SelectInstrumentPrice{
							Category: obS.Category,
							Symbol:   obS.Symbol,
							Price:    obS.Asks[askIdx][0],
						}
					}
				} else {
					bidIdx := int((s.E.Mouse.Pos.Y - ob.ySep + 1) / ob.rh)
					s.BTX <- &core.SelectInstrumentPrice{
						Category: obS.Category,
						Symbol:   obS.Symbol,
						Price:    obS.Bids[bidIdx][0],
					}
				}
			}
		}

		s.E.Mouse.Captured = true
	}

	rl.EndScissorMode()

	ob.Outline(1, s.P.Bg[0])
}

func (ob *OrderBook) RenderSplitView(s *core.State, obS *core.OrderBookState) {
	n := min(len(obS.Bids), len(obS.Asks))
	if n == 0 {
		return
	}

	ob.rh = s.RH - obS.RHD

	halfX := ob.s.X * .5

	if s.WRF || obS.Forced {
		ob.xSep = ob.p.X + halfX
		offsetY := min(0, obS.OffsetY) * -1
		ob.cap = min(n, int((ob.s.Y+offsetY)/ob.rh)+1)

		ob.UpdateOrderBookState(s, obS)
	}

	rl.BeginScissorMode(
		int32(ob.p.X),
		int32(ob.p.Y),
		int32(ob.s.X),
		int32(ob.s.Y),
	)

	maxSizeW := halfX - 1

	for i := range ob.cap {
		offsetY := float32(i)*ob.rh + obS.OffsetY

		bid := obS.Bids[i]
		bidSizeRatio := float32(bid[1] / ob.maxBidS)
		bidSizeW := maxSizeW * bidSizeRatio

		bidPos := rl.Vector2{
			X: ob.xSep - bidSizeW - 2,
			Y: ob.p.Y + offsetY,
		}

		rl.DrawRectangleV(
			bidPos,
			rl.Vector2{X: bidSizeW, Y: ob.rh},
			s.P.Diff.Add,
		)

		rl.DrawTextEx(
			s.F,
			ob.bidsText[i][0],
			rl.Vector2{
				X: ob.xSep - ORDER_BOOK_TEXT_XPD - ob.bidsPriceTextW[i],
				Y: bidPos.Y,
			},
			ob.rh,
			0,
			s.P.Fg[1],
		)

		rl.DrawLineEx(
			rl.Vector2{X: ob.p.X, Y: bidPos.Y + ob.rh},
			rl.Vector2{X: ob.xSep, Y: bidPos.Y + ob.rh},
			1,
			s.P.Bg[0],
		)

		ask := obS.Asks[i]
		askSizeRatio := float32(ask[1] / ob.maxAskS)
		askSizeW := halfX * askSizeRatio

		askPos := rl.Vector2{
			X: ob.xSep + 2,
			Y: ob.p.Y + offsetY,
		}

		rl.DrawRectangleV(
			askPos,
			rl.Vector2{X: askSizeW, Y: ob.rh},
			s.P.Diff.Delete,
		)

		rl.DrawTextEx(
			s.F,
			ob.asksText[i][0],
			rl.Vector2{
				X: ob.xSep + ORDER_BOOK_TEXT_XPD,
				Y: askPos.Y,
			},
			ob.rh,
			0,
			s.P.Fg[1],
		)

		rl.DrawLineEx(
			rl.Vector2{X: ob.xSep, Y: askPos.Y + ob.rh},
			rl.Vector2{X: ob.p.X + ob.s.X, Y: askPos.Y + ob.rh},
			1,
			s.P.Bg[0],
		)
	}
	// Vertical center line
	rl.DrawLineEx(
		rl.Vector2{X: ob.xSep, Y: ob.p.Y},
		rl.Vector2{X: ob.xSep, Y: ob.p.Y + ob.s.Y},
		2,
		s.P.Base.Orange,
	)

	if !s.E.Mouse.Captured && ob.ContainsV(s.E.Mouse.Pos) {

		if s.E.Mouse.HoldLeft {
			obS.OffsetY -= s.E.Mouse.Delta.Y
		} else {
			if s.E.Mouse.Click[0] {
				idx := int((s.E.Mouse.Pos.Y - ob.p.Y) / ob.rh)
				if s.E.Mouse.Pos.X > ob.xSep {
					if len(obS.Asks) > idx {
						s.BTX <- &core.SelectInstrumentPrice{
							Category: obS.Category,
							Symbol:   obS.Symbol,
							Price:    obS.Asks[idx][0],
						}
					}
				} else {
					if len(obS.Asks) > idx {
						s.BTX <- &core.SelectInstrumentPrice{
							Category: obS.Category,
							Symbol:   obS.Symbol,
							Price:    obS.Bids[idx][0],
						}
					}
				}
			}
		}

		s.E.Mouse.Captured = true
	}

	rl.EndScissorMode()

	ob.Outline(1, s.P.Bg[0])
}

func (ob *OrderBook) UpdateOrderBookState(s *core.State, obS *core.OrderBookState) {

	ob.maxBidS = 0
	ob.maxAskS = 0

	ob.bidsPriceTextW = ob.bidsPriceTextW[:0]
	ob.asksPriceTextW = ob.asksPriceTextW[:0]

	ob.bidsText = ob.bidsText[:0]
	ob.asksText = ob.asksText[:0]

	rhScale := ob.rh / float32(s.F.BaseSize)

	for i := range ob.cap {
		bid := obS.Bids[i]
		bidPrice, bidSize := bid[0], bid[1]

		if bidSize > ob.maxBidS {
			ob.maxBidS = bidSize
		}

		bidPriceText := strconv.FormatFloat(bidPrice, 'f', -1, 64)

		var bidPriceTextW float32
		if bidPrice != float64(int(bidPrice)) {
			bidPriceTextW = s.TextNumSV.X*
				float32(len(bidPriceText)-1)*rhScale +
				s.TextDotW*rhScale
		} else {
			bidPriceTextW = s.TextNumSV.X *
				float32(len(bidPriceText)) * rhScale
		}

		ob.bidsPriceTextW = append(
			ob.bidsPriceTextW, bidPriceTextW,
		)

		bidSizeText := strconv.FormatFloat(bidSize, 'f', -1, 64)

		ob.bidsText = append(
			ob.bidsText, [2]string{bidPriceText, bidSizeText},
		)

		ask := obS.Asks[i]
		askPrice, askSize := ask[0], ask[1]

		if askSize > ob.maxAskS {
			ob.maxAskS = askSize
		}

		askPriceText := strconv.FormatFloat(askPrice, 'f', -1, 64)

		var askPriceTextW float32
		if askPrice != float64(int(askPrice)) {
			askPriceTextW = s.TextNumSV.X*
				float32(len(askPriceText)-1)*rhScale +
				s.TextDotW*rhScale
		} else {
			askPriceTextW = s.TextNumSV.X *
				float32(len(askPriceText)) * rhScale
		}

		ob.asksPriceTextW = append(
			ob.asksPriceTextW, askPriceTextW,
		)

		askSizeText := strconv.FormatFloat(askSize, 'f', -1, 64)

		ob.asksText = append(
			ob.asksText, [2]string{askPriceText, askSizeText},
		)
	}

}

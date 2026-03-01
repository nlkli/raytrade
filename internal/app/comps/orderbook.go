package comps

import (
	"nlkli/raytrade/internal/app/core"
	"strconv"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	ORDER_BOOK_TEXT_XPD float32 = 4 // Padding
)

type OrderBookPlus struct {
	splitter *Splitter
}

func (ob *OrderBookPlus) R() *Rect {
	return ob.splitter.Rect
}

func (ob *OrderBookPlus) Render(s *core.State) {
	ob.splitter.Render(s)
}

type OrderBook struct {
	*Rect

	StateIdx int

	VM       int
	ShowText bool

	RH float32

	ySep float32
	xSep float32

	Cap int

	MaxBidS,
	MaxAskS float64

	BidsText [][2]string
	AsksText [][2]string

	BidsPriceTextW []float32
	AsksPriceTextW []float32
}

func (ob *OrderBook) R() *Rect {
	return ob.Rect
}

func (ob *OrderBook) Render(s *core.State) {
	obS := s.OrderBook[ob.StateIdx]

	switch ob.VM {
	case 0:
		ob.RenderCenteredView(s, obS)
	default:
		ob.RenderSplitView(s, obS)
	}

	if obS.Forced {
		switch obS.PlusCompI {
		case 0:
			obS.Forced = false
		case 1:
			obS.PlusCompI++
		case 2:
			obS.PlusCompI--
			obS.Forced = false
		}
	}
}

func (ob *OrderBook) RenderCenteredView(s *core.State, obS *core.OrderBookState) {
	n := min(len(obS.Bids), len(obS.Asks))
	if n == 0 {
		return
	}

	ob.RH = s.RH - obS.RHD

	halfY := ob.s.Y * .5
	ob.ySep = ob.p.Y + halfY + obS.ShiftY

	if s.WRF || obS.Forced {
		shiftY := obS.ShiftY
		if shiftY < 0 {
			shiftY = -shiftY
		}
		ob.Cap = min(n, int((halfY+shiftY)/ob.RH)+1)
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

	if ob.ShowText {

		for i := range ob.Cap {
			offsetY := 2 + float32(i)*ob.RH

			bid := obS.Bids[i]

			bidSizeRatio := float32(bid[1] / ob.MaxBidS)
			bidSizeW := maxSizeW * bidSizeRatio

			bidPos := rl.Vector2{X: ob.p.X, Y: ob.ySep + offsetY}

			rl.DrawRectangleV(
				bidPos,
				rl.Vector2{X: bidSizeW, Y: ob.RH},
				s.P.Diff.Add,
			)

			rl.DrawTextEx(
				s.F,
				ob.BidsText[i][1],
				rl.Vector2{
					X: bidPos.X + ORDER_BOOK_TEXT_XPD,
					Y: bidPos.Y,
				},
				ob.RH,
				0,
				s.P.Fg[2],
			)

			rl.DrawTextEx(
				s.F,
				ob.BidsText[i][0],
				rl.Vector2{
					X: ob.p.X + ob.s.X -
						ob.BidsPriceTextW[i] - ORDER_BOOK_TEXT_XPD,
					Y: bidPos.Y,
				},
				ob.RH,
				0,
				s.P.Fg[1],
			)

			rl.DrawLineEx(
				rl.Vector2{X: ob.p.X, Y: bidPos.Y + ob.RH},
				rl.Vector2{X: ob.p.X + ob.s.X, Y: bidPos.Y + ob.RH},
				1,
				s.P.Bg[0],
			)

			ask := obS.Asks[i]

			askSizeRatio := float32(ask[1] / ob.MaxAskS)
			askSizeW := ob.s.X * askSizeRatio

			askPos := rl.Vector2{X: ob.p.X, Y: ob.ySep - offsetY - ob.RH}

			rl.DrawRectangleV(
				askPos,
				rl.Vector2{X: askSizeW, Y: ob.RH},
				s.P.Diff.Delete,
			)

			rl.DrawTextEx(
				s.F,
				ob.AsksText[i][1],
				rl.Vector2{
					X: askPos.X + ORDER_BOOK_TEXT_XPD,
					Y: askPos.Y,
				},
				ob.RH,
				0,
				s.P.Fg[2],
			)

			rl.DrawTextEx(
				s.F,
				ob.AsksText[i][0],
				rl.Vector2{
					X: ob.p.X + ob.s.X -
						ob.AsksPriceTextW[i] - ORDER_BOOK_TEXT_XPD,
					Y: askPos.Y,
				},
				ob.RH,
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

	} else {

		for i := range ob.Cap {
			offsetY := 2 + float32(i)*ob.RH

			bid := obS.Bids[i]

			bidSizeRatio := float32(bid[1] / ob.MaxBidS)
			bidSizeW := maxSizeW * bidSizeRatio

			bidPos := rl.Vector2{X: ob.p.X, Y: ob.ySep + offsetY}

			rl.DrawRectangleV(
				bidPos,
				rl.Vector2{X: bidSizeW, Y: ob.RH},
				s.P.Diff.Add,
			)

			rl.DrawLineEx(
				rl.Vector2{X: ob.p.X, Y: bidPos.Y + ob.RH},
				rl.Vector2{X: ob.p.X + ob.s.X, Y: bidPos.Y + ob.RH},
				1,
				s.P.Bg[0],
			)

			ask := obS.Asks[i]

			askSizeRatio := float32(ask[1] / ob.MaxAskS)
			askSizeW := ob.s.X * askSizeRatio

			askPos := rl.Vector2{X: ob.p.X, Y: ob.ySep - offsetY - ob.RH}

			rl.DrawRectangleV(
				askPos,
				rl.Vector2{X: askSizeW, Y: ob.RH},
				s.P.Diff.Delete,
			)

			rl.DrawLineEx(
				rl.Vector2{X: ob.p.X, Y: askPos.Y},
				rl.Vector2{X: ob.p.X + ob.s.X, Y: askPos.Y},
				1,
				s.P.Bg[0],
			)
		}

	}

	if !s.E.Mouse.Captured && ob.ContainsV(s.E.Mouse.Pos) {

		if s.E.Mouse.HoldLeft {
			obS.ShiftY -= s.E.Mouse.Delta.Y
		} else {
			if s.E.Mouse.Click[0] {
				if s.E.Mouse.Pos.Y < ob.ySep {
					askIdx := int((ob.ySep - s.E.Mouse.Pos.Y - 1) / ob.RH)
					if len(obS.Asks) > askIdx {
						s.Select.Price.InstrumentKey = obS.InstrumentKey
						s.Select.Price.Value = obS.Asks[askIdx][0]
					}
				} else {
					bidIdx := int((s.E.Mouse.Pos.Y - ob.ySep + 1) / ob.RH)
					if len(obS.Asks) > bidIdx {
						s.Select.Price.InstrumentKey = obS.InstrumentKey
						s.Select.Price.Value = obS.Bids[bidIdx][0]
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

	ob.RH = s.RH - obS.RHD

	halfX := ob.s.X * .5

	if s.WRF || obS.Forced {
		ob.xSep = ob.p.X + halfX
		offsetY := min(0, obS.OffsetY) * -1
		ob.Cap = min(n, int((ob.s.Y+offsetY)/ob.RH)+1)

		ob.UpdateOrderBookState(s, obS)
	}

	rl.BeginScissorMode(
		int32(ob.p.X),
		int32(ob.p.Y),
		int32(ob.s.X),
		int32(ob.s.Y),
	)

	maxSizeW := halfX - 1

	if ob.ShowText {
		for i := range ob.Cap {
			offsetY := float32(i)*ob.RH + obS.OffsetY

			bid := obS.Bids[i]
			bidSizeRatio := float32(bid[1] / ob.MaxBidS)
			bidSizeW := maxSizeW * bidSizeRatio

			bidPos := rl.Vector2{
				X: ob.xSep - bidSizeW - 2,
				Y: ob.p.Y + offsetY,
			}

			rl.DrawRectangleV(
				bidPos,
				rl.Vector2{X: bidSizeW, Y: ob.RH},
				s.P.Diff.Add,
			)

			rl.DrawTextEx(
				s.F,
				ob.BidsText[i][0],
				rl.Vector2{
					X: ob.xSep - ORDER_BOOK_TEXT_XPD - ob.BidsPriceTextW[i],
					Y: bidPos.Y,
				},
				ob.RH,
				0,
				s.P.Fg[1],
			)

			rl.DrawLineEx(
				rl.Vector2{X: ob.p.X, Y: bidPos.Y + ob.RH},
				rl.Vector2{X: ob.xSep, Y: bidPos.Y + ob.RH},
				1,
				s.P.Bg[0],
			)

			ask := obS.Asks[i]
			askSizeRatio := float32(ask[1] / ob.MaxAskS)
			askSizeW := halfX * askSizeRatio

			askPos := rl.Vector2{
				X: ob.xSep + 2,
				Y: ob.p.Y + offsetY,
			}

			rl.DrawRectangleV(
				askPos,
				rl.Vector2{X: askSizeW, Y: ob.RH},
				s.P.Diff.Delete,
			)

			rl.DrawTextEx(
				s.F,
				ob.AsksText[i][0],
				rl.Vector2{
					X: ob.xSep + ORDER_BOOK_TEXT_XPD,
					Y: askPos.Y,
				},
				ob.RH,
				0,
				s.P.Fg[1],
			)

			rl.DrawLineEx(
				rl.Vector2{X: ob.xSep, Y: askPos.Y + ob.RH},
				rl.Vector2{X: ob.p.X + ob.s.X, Y: askPos.Y + ob.RH},
				1,
				s.P.Bg[0],
			)
		}
	} else {
		for i := range ob.Cap {
			offsetY := float32(i)*ob.RH + obS.OffsetY

			bid := obS.Bids[i]
			bidSizeRatio := float32(bid[1] / ob.MaxBidS)
			bidSizeW := maxSizeW * bidSizeRatio

			bidPos := rl.Vector2{
				X: ob.xSep - bidSizeW - 2,
				Y: ob.p.Y + offsetY,
			}

			rl.DrawRectangleV(
				bidPos,
				rl.Vector2{X: bidSizeW, Y: ob.RH},
				s.P.Diff.Add,
			)

			rl.DrawLineEx(
				rl.Vector2{X: ob.p.X, Y: bidPos.Y + ob.RH},
				rl.Vector2{X: ob.xSep, Y: bidPos.Y + ob.RH},
				1,
				s.P.Bg[0],
			)

			ask := obS.Asks[i]
			askSizeRatio := float32(ask[1] / ob.MaxAskS)
			askSizeW := halfX * askSizeRatio

			askPos := rl.Vector2{
				X: ob.xSep + 2,
				Y: ob.p.Y + offsetY,
			}

			rl.DrawRectangleV(
				askPos,
				rl.Vector2{X: askSizeW, Y: ob.RH},
				s.P.Diff.Delete,
			)

			rl.DrawLineEx(
				rl.Vector2{X: ob.xSep, Y: askPos.Y + ob.RH},
				rl.Vector2{X: ob.p.X + ob.s.X, Y: askPos.Y + ob.RH},
				1,
				s.P.Bg[0],
			)
		}
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
				idx := int((s.E.Mouse.Pos.Y - ob.p.Y) / ob.RH)
				if s.E.Mouse.Pos.X > ob.xSep {
					if len(obS.Asks) > idx {
						s.Select.Price.InstrumentKey = obS.InstrumentKey
						s.Select.Price.Value = obS.Asks[idx][0]
					}
				} else {
					if len(obS.Asks) > idx {
						s.Select.Price.InstrumentKey = obS.InstrumentKey
						s.Select.Price.Value = obS.Bids[idx][0]
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

	ob.MaxBidS = 0
	ob.MaxAskS = 0

	if ob.ShowText {

		ob.BidsPriceTextW = ob.BidsPriceTextW[:0]
		ob.AsksPriceTextW = ob.AsksPriceTextW[:0]

		ob.BidsText = ob.BidsText[:0]
		ob.AsksText = ob.AsksText[:0]

		rhScale := ob.RH / float32(s.F.BaseSize)

		for i := range ob.Cap {
			bid := obS.Bids[i]
			bidPrice, bidSize := bid[0], bid[1]

			if bidSize > ob.MaxBidS {
				ob.MaxBidS = bidSize
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

			ob.BidsPriceTextW = append(
				ob.BidsPriceTextW, bidPriceTextW,
			)

			bidSizeText := strconv.FormatFloat(bidSize, 'f', -1, 64)

			ob.BidsText = append(
				ob.BidsText, [2]string{bidPriceText, bidSizeText},
			)

			ask := obS.Asks[i]
			askPrice, askSize := ask[0], ask[1]

			if askSize > ob.MaxAskS {
				ob.MaxAskS = askSize
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

			ob.AsksPriceTextW = append(
				ob.AsksPriceTextW, askPriceTextW,
			)

			askSizeText := strconv.FormatFloat(askSize, 'f', -1, 64)

			ob.AsksText = append(
				ob.AsksText, [2]string{askPriceText, askSizeText},
			)
		}

	} else {

		for i := range ob.Cap {
			bidSize := obS.Bids[i][1]

			if bidSize > ob.MaxBidS {
				ob.MaxBidS = bidSize
			}

			askSize := obS.Asks[i][1]

			if askSize > ob.MaxAskS {
				ob.MaxAskS = askSize
			}
		}

	}
}

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

	CenterY float32
	CenterX float32

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

	if s.WRF || obS.Forced {
		halfY := ob.s.Y * .5

		ob.CenterY = ob.p.Y + halfY
		ob.Cap = min(n, int(halfY/ob.RH)+1)

		ob.UpdateOrderBookState(s, obS)
	}

	rl.BeginScissorMode(
		int32(ob.p.X),
		int32(ob.p.Y),
		int32(ob.s.X),
		int32(ob.s.Y),
	)

	rl.DrawLineEx(
		rl.Vector2{X: ob.p.X, Y: ob.CenterY},
		rl.Vector2{X: ob.p.X + ob.s.X, Y: ob.CenterY},
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

			bidPos := rl.Vector2{X: ob.p.X, Y: ob.CenterY + offsetY}

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

			askPos := rl.Vector2{X: ob.p.X, Y: ob.CenterY - offsetY - ob.RH}

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

			bidPos := rl.Vector2{X: ob.p.X, Y: ob.CenterY + offsetY}

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

			askPos := rl.Vector2{X: ob.p.X, Y: ob.CenterY - offsetY - ob.RH}

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
		ob.CenterX = ob.p.X + halfX
		ob.Cap = min(n, int(ob.s.Y/ob.RH)+1)

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
			offsetY := float32(i) * ob.RH

			bid := obS.Bids[i]
			bidSizeRatio := float32(bid[1] / ob.MaxBidS)
			bidSizeW := maxSizeW * bidSizeRatio

			bidPos := rl.Vector2{
				X: ob.CenterX - bidSizeW - 2,
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
					X: ob.CenterX - ORDER_BOOK_TEXT_XPD - ob.BidsPriceTextW[i],
					Y: bidPos.Y,
				},
				ob.RH,
				0,
				s.P.Fg[1],
			)

			rl.DrawLineEx(
				rl.Vector2{X: ob.p.X, Y: bidPos.Y + ob.RH},
				rl.Vector2{X: ob.CenterX, Y: bidPos.Y + ob.RH},
				1,
				s.P.Bg[0],
			)

			ask := obS.Asks[i]
			askSizeRatio := float32(ask[1] / ob.MaxAskS)
			askSizeW := halfX * askSizeRatio

			askPos := rl.Vector2{
				X: ob.CenterX + 2,
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
					X: ob.CenterX + ORDER_BOOK_TEXT_XPD,
					Y: askPos.Y,
				},
				ob.RH,
				0,
				s.P.Fg[1],
			)

			rl.DrawLineEx(
				rl.Vector2{X: ob.CenterX, Y: askPos.Y + ob.RH},
				rl.Vector2{X: ob.p.X + ob.s.X, Y: askPos.Y + ob.RH},
				1,
				s.P.Bg[0],
			)
		}
	} else {
		for i := range ob.Cap {
			offsetY := float32(i) * ob.RH

			bid := obS.Bids[i]
			bidSizeRatio := float32(bid[1] / ob.MaxBidS)
			bidSizeW := maxSizeW * bidSizeRatio

			bidPos := rl.Vector2{
				X: ob.CenterX - bidSizeW - 2,
				Y: ob.p.Y + offsetY,
			}

			rl.DrawRectangleV(
				bidPos,
				rl.Vector2{X: bidSizeW, Y: ob.RH},
				s.P.Diff.Add,
			)

			rl.DrawLineEx(
				rl.Vector2{X: ob.p.X, Y: bidPos.Y + ob.RH},
				rl.Vector2{X: ob.CenterX, Y: bidPos.Y + ob.RH},
				1,
				s.P.Bg[0],
			)

			ask := obS.Asks[i]
			askSizeRatio := float32(ask[1] / ob.MaxAskS)
			askSizeW := halfX * askSizeRatio

			askPos := rl.Vector2{
				X: ob.CenterX + 2,
				Y: ob.p.Y + offsetY,
			}

			rl.DrawRectangleV(
				askPos,
				rl.Vector2{X: askSizeW, Y: ob.RH},
				s.P.Diff.Delete,
			)

			rl.DrawLineEx(
				rl.Vector2{X: ob.CenterX, Y: askPos.Y + ob.RH},
				rl.Vector2{X: ob.p.X + ob.s.X, Y: askPos.Y + ob.RH},
				1,
				s.P.Bg[0],
			)
		}
	}

	// Vertical center line
	rl.DrawLineEx(
		rl.Vector2{X: ob.CenterX, Y: ob.p.Y},
		rl.Vector2{X: ob.CenterX, Y: ob.p.Y + ob.s.Y},
		2,
		s.P.Base.Orange,
	)

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

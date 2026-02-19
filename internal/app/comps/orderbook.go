package comps

import (
	"nlkli/raytrade/internal/app/core"
	"strconv"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	ORDER_BOOK_WIDTH    float32 = 220
	ORDER_BOOK_RHL      int     = 1
	ORDER_BOOK_FILL_XPD float32 = 4 // Padding
)

type OrderBook struct {
	*Rect

	i int

	centerY float32
}

func (ob *OrderBook) R() *Rect {
	return ob.Rect
}

func (ob *OrderBook) Render(s *core.State) {
	obS := s.OrderBook[ob.i]

	n := min(len(obS.Bids), len(obS.Asks))
	if n == 0 {
		return
	}

	rh := s.RHL(1)

	if s.WRF || obS.Forced || s.RH_Dirty {
		ob.centerY = ob.p.Y + ob.s.Y*0.5

		obS.Cap = min(n, int(ob.s.Y*.5/rh)+1)

		obS.MaxBidS = 0
		obS.MaxAskS = 0

		obS.BidsPriceTextW = obS.BidsPriceTextW[:0]
		obS.AsksPriceTextW = obS.AsksPriceTextW[:0]

		obS.BidsText = obS.BidsText[:0]
		obS.AsksText = obS.AsksText[:0]

		rhlScale := s.RHL_Scale(ORDER_BOOK_RHL)

		for i := range obS.Cap {
			bid := obS.Bids[i]

			if bid[1] > obS.MaxBidS {
				obS.MaxBidS = bid[1]
			}

			bidPriceText := strconv.FormatFloat(bid[0], 'f', -1, 64)

			var bidPriceTextW float32
			if bid[0] != float64(int(bid[0])) {
				bidPriceTextW = s.TextNumSV.X*
					float32(len(bidPriceText)-1)*rhlScale +
					s.TextDotW*rhlScale
			} else {
				bidPriceTextW = s.TextNumSV.X *
					float32(len(bidPriceText)) * rhlScale
			}

			obS.BidsPriceTextW = append(
				obS.BidsPriceTextW, bidPriceTextW,
			)

			bidSizeText := strconv.FormatFloat(bid[1], 'f', -1, 64)

			obS.BidsText = append(
				obS.BidsText, [2]string{bidPriceText, bidSizeText},
			)

			ask := obS.Asks[i]

			if ask[1] > obS.MaxAskS {
				obS.MaxAskS = ask[1]
			}

			askPriceText := strconv.FormatFloat(ask[0], 'f', -1, 64)

			var askPriceTextW float32
			if ask[0] != float64(int(ask[0])) {
				askPriceTextW = s.TextNumSV.X*
					float32(len(askPriceText)-1)*rhlScale +
					s.TextDotW*rhlScale
			} else {
				askPriceTextW = s.TextNumSV.X *
					float32(len(askPriceText)) * rhlScale
			}

			obS.AsksPriceTextW = append(
				obS.AsksPriceTextW, askPriceTextW,
			)

			askSizeText := strconv.FormatFloat(ask[1], 'f', -1, 64)

			obS.AsksText = append(
				obS.AsksText, [2]string{askPriceText, askSizeText},
			)
		}

		if obS.Forced {
			obS.Forced = false
		}

		// obS.Forced = false
	}

	rl.BeginScissorMode(
		int32(ob.p.X),
		int32(ob.p.Y),
		int32(ob.s.X),
		int32(ob.s.Y),
	)

	rl.DrawLineEx(
		rl.Vector2{X: ob.p.X, Y: ob.centerY},
		rl.Vector2{X: ob.p.X + ob.s.X, Y: ob.centerY},
		2,
		s.P.Base.Orange,
	)

	xPos := ob.p.X + ORDER_BOOK_FILL_XPD

	for i := range obS.Cap {
		offsetY := 2 + float32(i)*rh

		ask := obS.Asks[i]

		askSizeRatio := float32(ask[1] / obS.MaxAskS)
		askSizeW := ob.s.X * askSizeRatio

		askPos := rl.Vector2{X: xPos, Y: ob.centerY - offsetY - rh}

		rl.DrawRectangleV(
			askPos,
			rl.Vector2{X: askSizeW, Y: rh},
			s.P.Diff.Delete,
		)

		rl.DrawTextEx(
			s.F,
			obS.AsksText[i][1],
			askPos,
			rh,
			0,
			s.P.Fg[2],
		)

		rl.DrawTextEx(
			s.F,
			obS.AsksText[i][0],
			rl.Vector2{
				X: ob.p.X + ob.s.X -
					ORDER_BOOK_FILL_XPD - obS.AsksPriceTextW[i],
				Y: askPos.Y,
			},
			rh,
			0,
			s.P.Fg[1],
		)

		rl.DrawLineEx(
			rl.Vector2{X: ob.p.X, Y: askPos.Y},
			rl.Vector2{X: ob.p.X + ob.s.X, Y: askPos.Y},
			1,
			s.P.Bg[0],
		)

		bid := obS.Bids[i]

		bidSizeRatio := float32(bid[1] / obS.MaxBidS)
		bidSizeW := ob.s.X * bidSizeRatio

		bidPos := rl.Vector2{X: xPos, Y: ob.centerY + offsetY}

		rl.DrawRectangleV(
			bidPos,
			rl.Vector2{X: bidSizeW, Y: rh},
			s.P.Diff.Add,
		)

		rl.DrawTextEx(
			s.F,
			obS.BidsText[i][1],
			bidPos,
			rh,
			0,
			s.P.Fg[2],
		)

		rl.DrawTextEx(
			s.F,
			obS.BidsText[i][0],
			rl.Vector2{
				X: ob.p.X + ob.s.X -
					ORDER_BOOK_FILL_XPD - obS.BidsPriceTextW[i],
				Y: bidPos.Y,
			},
			rh,
			0,
			s.P.Fg[1],
		)

		rl.DrawLineEx(
			rl.Vector2{X: ob.p.X, Y: bidPos.Y + rh},
			rl.Vector2{X: ob.p.X + ob.s.X, Y: bidPos.Y + rh},
			1,
			s.P.Bg[0],
		)
	}

	rl.EndScissorMode()

	ob.Outline(1, s.P.Bg[0])
}

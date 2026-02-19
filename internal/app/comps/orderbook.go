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

func CreateOrderBookComponent(params map[string]any) Comp {
	return &OrderBook{
		Rect: &Rect{},
	}
}

type OrderBook struct {
	*Rect

	centerY float32
}

func (ob *OrderBook) R() *Rect {
	return ob.Rect
}

func (ob *OrderBook) Render(s *core.State) {
	n := min(len(s.OrderBook.Bids), len(s.OrderBook.Asks))
	if n == 0 {
		return
	}

	rh := s.RHL(1)

	if s.WRF || s.OrderBook.Forced || s.RH_Dirty {
		ob.centerY = ob.p.Y + ob.s.Y*0.5

		s.OrderBook.Cap = min(n, int(ob.s.Y*.5/rh)+1)

		s.OrderBook.MaxBidS = 0
		s.OrderBook.MaxAskS = 0

		s.OrderBook.BidsPriceTextW = s.OrderBook.BidsPriceTextW[:0]
		s.OrderBook.AsksPriceTextW = s.OrderBook.AsksPriceTextW[:0]

		s.OrderBook.BidsText = s.OrderBook.BidsText[:0]
		s.OrderBook.AsksText = s.OrderBook.AsksText[:0]

		rhlScale := s.RHL_Scale(ORDER_BOOK_RHL)

		for i := range s.OrderBook.Cap {
			bid := s.OrderBook.Bids[i]

			if bid[1] > s.OrderBook.MaxBidS {
				s.OrderBook.MaxBidS = bid[1]
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

			s.OrderBook.BidsPriceTextW = append(
				s.OrderBook.BidsPriceTextW, bidPriceTextW,
			)

			bidSizeText := strconv.FormatFloat(bid[1], 'f', -1, 64)

			s.OrderBook.BidsText = append(
				s.OrderBook.BidsText, [2]string{bidPriceText, bidSizeText},
			)

			ask := s.OrderBook.Asks[i]

			if ask[1] > s.OrderBook.MaxAskS {
				s.OrderBook.MaxAskS = ask[1]
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

			s.OrderBook.AsksPriceTextW = append(
				s.OrderBook.AsksPriceTextW, askPriceTextW,
			)

			askSizeText := strconv.FormatFloat(ask[1], 'f', -1, 64)

			s.OrderBook.AsksText = append(
				s.OrderBook.AsksText, [2]string{askPriceText, askSizeText},
			)
		}

		if s.OrderBook.Forced {
			s.OrderBook.Forced = false
		}

		// s.OrderBook.Forced = false
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

	for i := range s.OrderBook.Cap {
		offsetY := 2 + float32(i)*rh

		ask := s.OrderBook.Asks[i]

		askSizeRatio := float32(ask[1] / s.OrderBook.MaxAskS)
		askSizeW := ob.s.X * askSizeRatio

		askPos := rl.Vector2{X: xPos, Y: ob.centerY - offsetY - rh}

		rl.DrawRectangleV(
			askPos,
			rl.Vector2{X: askSizeW, Y: rh},
			s.P.Diff.Delete,
		)

		rl.DrawTextEx(
			s.F,
			s.OrderBook.AsksText[i][1],
			askPos,
			rh,
			0,
			s.P.Fg[2],
		)

		rl.DrawTextEx(
			s.F,
			s.OrderBook.AsksText[i][0],
			rl.Vector2{
				X: ob.p.X + ob.s.X -
					ORDER_BOOK_FILL_XPD - s.OrderBook.AsksPriceTextW[i],
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

		bid := s.OrderBook.Bids[i]

		bidSizeRatio := float32(bid[1] / s.OrderBook.MaxBidS)
		bidSizeW := ob.s.X * bidSizeRatio

		bidPos := rl.Vector2{X: xPos, Y: ob.centerY + offsetY}

		rl.DrawRectangleV(
			bidPos,
			rl.Vector2{X: bidSizeW, Y: rh},
			s.P.Diff.Add,
		)

		rl.DrawTextEx(
			s.F,
			s.OrderBook.BidsText[i][1],
			bidPos,
			rh,
			0,
			s.P.Fg[2],
		)

		rl.DrawTextEx(
			s.F,
			s.OrderBook.BidsText[i][0],
			rl.Vector2{
				X: ob.p.X + ob.s.X -
					ORDER_BOOK_FILL_XPD - s.OrderBook.BidsPriceTextW[i],
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

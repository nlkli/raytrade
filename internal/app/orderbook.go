package app

import (
	"strconv"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type OrderBook struct {
	*Rect
	parent *Rect

	centerY float32
}

func (ob *OrderBook) Render(s *State) {
	if s.WRF || s.OrderBook.Forced {
		ob.centerY = ob.p.Y + ob.s.Y*0.5
	}

	n := min(len(s.OrderBook.Bids), len(s.OrderBook.Asks))
	if n == 0 {
		return
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

	rh := s.RHL(1)
	// gap := rh * .5
	// step := rh + gap

	for i := range n {
		offset := 2 + float32(i)*rh

		priceText := strconv.FormatFloat(s.OrderBook.Asks[i][0], 'f', -1, 64)

		aV := rl.Vector2{X: ob.p.X, Y: ob.centerY - offset - rh}

		rl.DrawTextEx(
			s.F,
			priceText,
			aV,
			rh,
			0,
			s.P.Fg[2],
		)

		rl.DrawLineEx(
			rl.Vector2{X: ob.p.X, Y: aV.Y},
			rl.Vector2{X: ob.p.X + ob.s.X, Y: aV.Y},
			1,
			s.P.Fg[3],
		)

		priceText = strconv.FormatFloat(s.OrderBook.Bids[i][0], 'f', -1, 64)

		bV := rl.Vector2{X: ob.p.X, Y: ob.centerY + offset}

		rl.DrawTextEx(
			s.F,
			priceText,
			bV,
			rh,
			0,
			s.P.Fg[2],
		)

		rl.DrawLineEx(
			rl.Vector2{X: ob.p.X, Y: bV.Y + rh},
			rl.Vector2{X: ob.p.X + ob.s.X, Y: bV.Y + rh},
			1,
			s.P.Fg[3],
		)
	}

	rl.EndScissorMode()
}

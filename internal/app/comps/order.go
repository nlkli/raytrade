package comps

import (
	"fmt"
	"nlkli/raytrade/internal/app/core"
	"nlkli/raytrade/internal/broker"
	"strconv"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	ORDER_CONTENT_PDX float32 = 4 // padding
)

type Order struct {
	*Rect
}

func (o *Order) R() *Rect {
	return o.Rect
}

func (o *Order) Render(s *core.State) {

	os := s.Order

	rh1 := s.RH - os.RHD

	if len(s.Order.List) == 0 {
		return
	}

	rl.BeginScissorMode(
		int32(o.p.X),
		int32(o.p.Y),
		int32(o.s.X),
		int32(o.s.Y),
	)

	var offsetY float32
	cursor := rl.Vector2{
		X: o.p.X + ORDER_CONTENT_PDX, Y: o.p.Y + offsetY,
	}
	for _, oi := range os.List {

		rl.DrawRectangleV(
			rl.Vector2{X: o.p.X, Y: cursor.Y},
			rl.Vector2{X: o.s.X, Y: s.RH},
			s.P.Bg[0],
		)

		rl.DrawTextEx(
			s.F,
			oi.Symbol,
			cursor,
			s.RH,
			0,
			s.P.Dim.Yellow,
		)

		cursor.Y += s.RH

		var sideText string
		if oi.Side == broker.Long {
			sideText = "Long"
		} else {
			sideText = "Short"
		}

		var sColor rl.Color
		if oi.Side == broker.Long {
			sColor = s.P.Dim.Green
		} else {
			sColor = s.P.Dim.Red
		}

		rl.DrawTextEx(
			s.F,
			sideText,
			cursor,
			rh1,
			0,
			sColor,
		)

		cursor.Y += rh1

		sizeText := fmt.Sprintf(
			"S: %s (%s)",
			strconv.FormatFloat(oi.LeavesQty, 'f', -1, 64),
			strconv.FormatFloat(oi.LeavesValue, 'f', -1, 64),
		)

		rl.DrawTextEx(
			s.F,
			sizeText,
			cursor,
			rh1,
			0,
			s.P.Fg[1],
		)

		cursor.Y += rh1

		priceText := fmt.Sprintf(
			"P: %s",
			strconv.FormatFloat(oi.Price, 'f', -1, 64),
		)

		rl.DrawTextEx(
			s.F,
			priceText,
			cursor,
			rh1,
			0,
			s.P.Fg[1],
		)

		cursor.Y += rh1
	}

	rl.DrawLineEx(
		rl.Vector2{X: o.p.X, Y: cursor.Y},
		rl.Vector2{X: o.p.X + o.s.X, Y: cursor.Y},
		1,
		s.P.Bg[0],
	)

	rl.EndScissorMode()

	o.Outline(1, s.P.Bg[0])
}

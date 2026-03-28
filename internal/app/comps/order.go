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

type orderForcedData struct {
	headerInfo string
	qtyInfo    string
	priceInfo  string
	height     float32
}

type Order struct {
	*Rect

	forcedData []orderForcedData
}

func (o *Order) R() *Rect {
	return o.Rect
}

func (o *Order) Render(s *core.State) {

	os := &s.Order

	rh1 := s.RH - os.RHD

	if len(s.Order.List) == 0 {
		return
	}

	if os.Forced {
		o.forcedData = o.forcedData[:0]

		for i, oi := range os.List {
			var fd orderForcedData

			fd.headerInfo = fmt.Sprintf(
				"%d.%s.%s", i, oi.Category.AsString(true), oi.Symbol,
			)

			fd.qtyInfo = fmt.Sprintf(
				"S: %s (%s)",
				strconv.FormatFloat(oi.LeavesQty, 'f', -1, 64),
				strconv.FormatFloat(oi.LeavesValue, 'f', -1, 64),
			)

			fd.priceInfo = fmt.Sprintf(
				"P: %s",
				strconv.FormatFloat(oi.Price, 'f', -1, 64),
			)

			fd.height = s.RH + rh1*3

			o.forcedData = append(o.forcedData, fd)
		}

		os.Forced = false
	}

	rl.BeginScissorMode(
		int32(o.p.X),
		int32(o.p.Y),
		int32(o.s.X),
		int32(o.s.Y),
	)

	cursor := rl.Vector2{
		X: o.p.X + ORDER_CONTENT_PDX, Y: o.p.Y + os.OffsetY,
	}
	for i, oi := range os.List {
		fd := o.forcedData[i]

		rl.DrawRectangleV(
			rl.Vector2{X: o.p.X, Y: cursor.Y},
			rl.Vector2{X: o.s.X, Y: s.RH},
			s.P.Bg[0],
		)

		rl.DrawTextEx(
			s.F,
			fd.headerInfo,
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

		rl.DrawTextEx(
			s.F,
			fd.qtyInfo,
			cursor,
			rh1,
			0,
			s.P.Fg[1],
		)

		cursor.Y += rh1

		rl.DrawTextEx(
			s.F,
			fd.priceInfo,
			cursor,
			rh1,
			0,
			s.P.Fg[1],
		)

		cursor.Y += rh1
	}

	if !s.E.Mouse.Captured && o.Rect.ContainsV(s.E.Mouse.Pos) {
		if s.E.Mouse.HoldLeft {
			os.OffsetY -= s.E.Mouse.Delta.Y
		}

		if s.E.Mouse.Click[0] && s.E.Mouse.Pos.Y >= o.p.Y+os.OffsetY {
			cursorY := o.p.Y + os.OffsetY
			for i := range os.List {
				fd := o.forcedData[i]
				cursorY += fd.height
				if s.E.Mouse.Pos.Y < cursorY {
					s.BTX <- &core.SelectOrder{
						Order: &os.List[i],
					}
					break
				}
			}
		}

		s.E.Mouse.Captured = true
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

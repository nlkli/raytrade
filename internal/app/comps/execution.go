package comps

import (
	"fmt"
	"nlkli/raytrade/internal/app/core"
	"nlkli/raytrade/internal/broker"
	"strconv"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type executionForcedData struct {
	headerInfo string
	exInfo     string
}

type Execution struct {
	*Rect

	rhScale    float32
	forcedData []executionForcedData
}

func (e *Execution) R() *Rect {
	return e.Rect
}

func (e *Execution) Render(s *core.State) {
	es := &s.Execution

	rh := s.RH - es.RHD

	if s.WRF || s.RH_Dirty {
		e.rhScale = rh / float32(s.F.BaseSize)
	}

	if len(es.List) == 0 {
		return
	}

	if es.Forced {
		e.forcedData = e.forcedData[:0]

		for i, ei := range es.List {
			var fd executionForcedData

			fd.headerInfo = fmt.Sprintf(
				"%d.%s.%s", i, ei.Category.AsString(true), ei.Symbol,
			)

			fd.exInfo = fmt.Sprintf(
				"%s | %s",
				strconv.FormatFloat(ei.Qty, 'f', -1, 64),
				strconv.FormatFloat(ei.Price, 'f', -1, 64),
			)

			e.forcedData = append(
				e.forcedData,
				fd,
			)
		}

		es.Forced = false
	}

	rl.BeginScissorMode(
		int32(e.p.X),
		int32(e.p.Y),
		int32(e.s.X),
		int32(e.s.Y),
	)

	cursor := rl.Vector2{
		X: e.p.X + POSITION_CONTENT_PDX, Y: e.p.Y + es.OffsetY,
	}

	for i := len(es.List) - 1; i >= 0; i-- {
		ei := es.List[i]

		fd := e.forcedData[i]

		rl.DrawRectangleV(
			rl.Vector2{X: e.p.X, Y: cursor.Y},
			rl.Vector2{X: e.s.X, Y: rh},
			s.P.Bg[0],
		)

		rl.DrawTextEx(
			s.F,
			fd.headerInfo,
			cursor,
			rh,
			0,
			s.P.Fg[1],
		)

		cursor.Y += rh

		var sideText string
		if ei.Side == broker.Long {
			sideText = "Long"
		} else {
			sideText = "Short"
		}

		var sColor rl.Color
		if ei.Side == broker.Long {
			sColor = s.P.Dim.Green
		} else {
			sColor = s.P.Dim.Red
		}

		rl.DrawTextEx(
			s.F,
			sideText,
			cursor,
			rh,
			0,
			sColor,
		)

		cursor.Y += rh

		rl.DrawTextEx(
			s.F,
			fd.exInfo,
			cursor,
			rh,
			0,
			s.P.Fg[1],
		)

		cursor.Y += rh
	}

	if !s.E.Mouse.Captured && e.Rect.ContainsV(s.E.Mouse.Pos) {
		if s.E.Mouse.HoldLeft {
			es.OffsetY -= s.E.Mouse.Delta.Y
		}

		s.E.Mouse.Captured = true
	}

	rl.DrawLineEx(
		rl.Vector2{X: e.p.X, Y: cursor.Y},
		rl.Vector2{X: e.p.X + e.s.X, Y: cursor.Y},
		1,
		s.P.Bg[0],
	)

	rl.EndScissorMode()

	e.Outline(1, s.P.Bg[0])
}

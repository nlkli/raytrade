package comps

import (
	"fmt"
	"nlkli/raytrade/internal/app/core"
	"nlkli/raytrade/internal/broker"
	"strconv"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type Position struct {
	*Rect
}

func (p *Position) R() *Rect {
	return p.Rect
}

func (p *Position) Render(s *core.State) {
	ps := s.Position

	if len(ps.List) == 0 {
		return
	}

	rl.BeginScissorMode(
		int32(p.p.X),
		int32(p.p.Y),
		int32(p.s.X),
		int32(p.s.Y),
	)

	rh := s.RH - ps.RHD[0]
	rh1 := s.RH - ps.RHD[1]

	for i, pi := range ps.List {
		offsetY := (rh + rh1*2) * float32(i)
		cursor := rl.Vector2{
			X: p.p.X, Y: p.p.Y + offsetY,
		}

		rl.DrawRectangleV(
			cursor,
			rl.Vector2{X: p.s.X, Y: rh},
			s.P.Bg[0],
		)

		rl.DrawTextEx(
			s.F,
			pi.Symbol,
			cursor,
			rh,
			0,
			s.P.Dim.Yellow,
		)

		cursor.Y += rh1

		sizeInfo := fmt.Sprintf("%v %v", pi.Size, pi.Side)
		var color rl.Color
		if pi.Side == broker.Long {
			color = s.P.Dim.Green
		} else {
			color = s.P.Dim.Red
		}
		rl.DrawTextEx(
			s.F,
			sizeInfo,
			cursor,
			rh1,
			0,
			color,
		)

		cursor.Y += rh1

		rl.DrawTextEx(
			s.F,
			strconv.FormatFloat(pi.EntryPrice, 'f', -1, 64),
			cursor,
			rh1,
			0,
			s.P.Fg[1],
		)
	}

	rl.EndScissorMode()

	p.Outline(1, s.P.Bg[0])
}

package comps

import (
	"nlkli/raytrade/internal/app/core"

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

	for i, pi := range ps.List {
		offsetY := s.RH * float32(i)
		rl.DrawTextEx(
			s.F,
			pi.Symbol,
			rl.Vector2{X: p.p.X, Y: p.p.Y + offsetY},
			s.RH,
			0,
			s.P.Base.Yellow,
		)
	}

}

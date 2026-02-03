package app2

import rl "github.com/gen2brain/raylib-go/raylib"

const (
	ROOT_PD  float32 = 4
	RH       float32 = 16       // Row height
	CLH      float32 = RH * 1.2 // Command line height
	FOOTER_H float32 = RH + CLH
)

type Root struct {
	*Rect
	parent *Rect

	f *Footer
}

func InitRoot() *Root {
	root := &Root{
		Rect: &Rect{},
	}
	root.f = &Footer{
		Rect:   &Rect{},
		parent: root.Rect,
	}
	root.f.sl = &StatusLine{
		Rect:   &Rect{},
		parent: root.f.Rect,
	}
	root.f.cl = &CommandLine{
		Rect:   &Rect{},
		parent: root.f.Rect,
	}

	return root
}

func (r *Root) Render(s *State) {
	if s.IsWindowResized() {
		r.MoveTo(ROOT_PD, ROOT_PD)
		r.SetSize(s.W.X-ROOT_PD*2, s.W.Y-ROOT_PD*2)
	}
	rl.ClearBackground(s.P.Bg[1])
	r.f.Render(s)
}

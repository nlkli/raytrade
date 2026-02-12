package app

import (
	"fmt"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type Root struct {
	*Rect
	parent *Rect

	// tl *TabsLine
	mc *MainContent
	f  *Footer
}

func InitRoot() *Root {
	root := &Root{
		Rect: &Rect{},
	}
	root.mc = &MainContent{
		Rect:   &Rect{},
		parent: root.Rect,
	}
	root.mc.ch = &Chart{
		Rect:   &Rect{},
		parent: root.mc.Rect,
	}
	root.mc.ch.c = &Canvas{
		Rect:   &Rect{},
		parent: root.mc.ch.Rect,
	}
	root.mc.ch.tl = &TimeLine{
		Rect:   &Rect{},
		parent: root.mc.ch.Rect,
	}
	root.mc.ch.pb = &PriceBar{
		Rect:   &Rect{},
		parent: root.mc.ch.Rect,
	}
	root.mc.ch.cr = &Crossing{
		Rect:   &Rect{},
		parent: root.mc.ch.Rect,
	}
	root.mc.ob = &OrderBook{
		Rect:   &Rect{},
		parent: root.mc.Rect,
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
	rl.ClearBackground(s.P.Bg[1])

	if s.WRF {
		r.MoveTo(RPD, RPD)
		r.SetSize(s.WS.X-RPD*2, s.WS.Y-RPD*2)
	}

	s.CommandLine.LinesH = s.RH * float32(len(s.CommandLine.Lines))
	s.Footer.Height = s.RH*2 + CMD_LINE_MARGIN_BOTTOM + s.CommandLine.LinesH

	r.f.Render(s)
	r.mc.Render(s)

	if s.ShowFPS {
		fps := fmt.Sprintf("%d", rl.GetFPS())
		s.StdDrawText(string(fps), rl.Vector2{X: RPD, Y: RPD}, s.P.Comment)
	}

	// if s.ShowOverlay {
	// 	fps := fmt.Sprintf("%d", rl.GetFPS())
	// 	s.StdDrawText(string(fps), rl.Vector2{X: RPD, Y: RPD}, s.P.Comment)
	// }
}

// type TabsLine struct {
// 	*Rect
// 	parent *Rect
// }

type MainContent struct {
	*Rect
	parent *Rect

	ch *Chart
	ob *OrderBook
}

func (mc *MainContent) Render(s *State) {
	if s.WRF {
		mc.MoveTo(mc.parent.p.X, mc.parent.p.Y)
		mc.SetSize(mc.parent.s.X, mc.parent.s.Y-s.Footer.Height)
	}

	mc.ch.Render(s)

	// mc.Outline(1, s.P.Base.Red)
}

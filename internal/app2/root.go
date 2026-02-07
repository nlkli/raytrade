package app2

import rl "github.com/gen2brain/raylib-go/raylib"

const (
	RPD      float32 = 4        // Root pading
	RH       float32 = 16       // Row height
	RH_I32   int32   = 16       // Row height int32
	CLH      float32 = RH * 1.2 // Command line height
	OBW      float32 = 200      // OrderBook section width
	TLH      float32 = 20       // Time line height
	PBW      float32 = 40       // Price bar width
	CW       float32 = 5        // Candle width
	CG       float32 = 2        // Candles gap
	CWW      float32 = 2        // Candle wick width
	FOOTER_H float32 = RH + CLH
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

	r.mc.Render(s)
	r.f.Render(s)
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
		mc.SetSize(mc.parent.s.X, mc.parent.s.Y-FOOTER_H)
	}

	mc.ch.Render(s)

	// mc.Outline(1, s.P.Base.Red)
}

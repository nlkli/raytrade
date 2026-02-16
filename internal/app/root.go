package app

import (
	"fmt"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type Root struct {
	*Rect

	mc MainContent
	f  Footer
}

func InitRoot() *Root {
	return &Root{
		Rect: &Rect{},
	}
}

func (r *Root) Render(s *State) {
	rl.ClearBackground(s.P.Bg[1])

	if s.WRF || s.RH_Dirty {
		r.MoveTo(RPD, RPD)
		r.SetSize(s.WS.X-RPD*2, s.WS.Y-RPD*2)

		footerH := s.RH*2 + CMD_LINE_MARGIN_BOTTOM +
			s.RH*float32(len(s.CommandLine.Lines))

		r.mc.Rect, r.f.Rect = r.SplitH(r.s.Y - footerH)
	}

	r.mc.Render(s)
	r.f.Render(s)

	if s.ShowFPS {
		fps := fmt.Sprintf("%d", rl.GetFPS())
		s.StdDrawText(string(fps), rl.Vector2{X: RPD, Y: RPD}, s.P.Comment)
	}

	if s.RH_Dirty {
		s.RH_Dirty = false
	}
}

type MainContent struct {
	*Rect

	ch Chart
	ob OrderBook
}

func (mc *MainContent) Render(s *State) {
	if s.WRF || s.RH_Dirty {
		mc.ch.Rect, mc.ob.Rect = mc.SplitV(mc.s.X - ORDER_BOOK_WIDTH)
	}

	mc.ch.Render(s)
	mc.ob.Render(s)

	// mc.Outline(1, s.P.Base.Red)
}

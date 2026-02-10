package app

import (
	"fmt"
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type Footer struct {
	*Rect
	parent *Rect

	sl *StatusLine
	cl *CommandLine
}

func (f *Footer) Render(s *State) {
	if s.WRF || s.Footer.Forced {
		fh := s.Footer.Height

		f.MoveTo(f.parent.p.X, f.parent.s.Y-fh+f.parent.p.Y)
		f.SetSize(f.parent.s.X, fh)
	}

	f.sl.Render(s)
	f.cl.Render(s)

	// f.Outline(1, s.P.Base.Pink)
}

type StatusLine struct {
	*Rect
	parent *Rect

	utS  string
	utTW float32
}

func (sl *StatusLine) Render(s *State) {
	if s.WRF || s.Footer.Forced {
		sl.MoveTo(sl.parent.p.X, sl.parent.p.Y)
		sl.SetSize(sl.parent.s.X, s.RH)
	}

	sl.Fill(s.P.Bg[0])

	if s.FN%uint64(s.TFPS/4) == 0 {
		sl.utS = time.Since(s.ST).Truncate(time.Second).String()
		sl.utTW = s.StdMeasureText(sl.utS).X
	}

	if len(sl.utS) > 0 {
		s.StdDrawText(sl.utS, rl.Vector2{X: sl.s.X - sl.utTW, Y: sl.p.Y}, s.P.Fg[3])
	}

	s.StdDrawText(
		// TODO interval to string
		fmt.Sprintf("%s/%s", s.StatusLine.Symbol, s.StatusLine.Interval),
		rl.Vector2{X: sl.p.X, Y: sl.p.Y},
		s.P.Fg[3],
	)
}

type CommandLine struct {
	*Rect
	parent *Rect
}

func (cl *CommandLine) Render(s *State) {

	if s.WRF || s.Footer.Forced {
		cl.MoveTo(cl.parent.p.X, cl.parent.p.Y+s.RH)
		cl.SetSize(cl.parent.s.X, s.RH+CMD_LINE_MARGIN_BOTTOM)

		s.Footer.Forced = false
	}

	if s.M == Input {
		rl.DrawRectangleV(
			rl.Vector2{
				X: cl.p.X + s.CommandLine.PromptW + 1,
				Y: cl.p.Y + s.CommandLine.LinesH,
			},
			rl.Vector2{X: s.RH / 2, Y: s.RH},
			s.P.Cur.Bg,
		)
	}

	if len(s.CommandLine.Lines) > 0 {
		for i, l := range s.CommandLine.Lines {
			height := s.RH * float32(i+1) // TODO limit
			s.StdDrawText(
				l,
				rl.Vector2{X: cl.p.X, Y: cl.parent.p.Y + height},
				s.CommandLine.Color,
			)
		}
	}

	if len(s.CommandLine.Prompt) > 0 {
		s.StdDrawText(
			s.CommandLine.Prompt,
			rl.Vector2{
				X: cl.p.X,
				Y: cl.p.Y + s.CommandLine.LinesH,
			},
			s.CommandLine.Color,
		)
	}

	// cl.Outline(1, s.P.Base.Pink)
}

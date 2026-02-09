package app

import (
	"fmt"
	"strings"
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
	if s.WRF || s.Cache.Static.FooterResizeF {
		fh := s.Cache.Static.FooterH

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
	if s.WRF || s.Cache.Static.FooterResizeF {
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
		fmt.Sprintf("%s/%s", s.StatusLine.Symbol, s.StatusLine.Interval.AsString()),
		rl.Vector2{X: sl.p.X, Y: sl.p.Y},
		s.P.Fg[3],
	)
}

type CommandLine struct {
	*Rect
	parent *Rect
}

func (cl *CommandLine) Render(s *State) {
	if s.WRF || s.Cache.Static.FooterResizeF {
		cl.MoveTo(cl.parent.p.X, cl.parent.p.Y+s.RH)
		cl.SetSize(cl.parent.s.X, s.RH+CMD_LINE_MARGIN_BOTTOM)

		s.Cache.Static.FooterResizeF = false
	}

	if s.M == Input {
		bspace := func() {
			if len(s.CommandLine.Prompt) > 1 {
				r := []rune(s.CommandLine.Prompt)
				s.CommandLine.Prompt = string(r[:len(r)-1])
			}
		}
		if rl.IsKeyPressed(rl.KeyBackspace) {
			bspace()
		} else {
			if s.FN%uint64(s.TFPS/4) == 0 {
				if rl.IsKeyDown(rl.KeyBackspace) {
					bspace()
				}
			}
		}

		tw := s.StdMeasureText(s.CommandLine.Prompt).X
		rl.DrawRectangleV(
			rl.Vector2{X: cl.p.X + tw + 1, Y: cl.p.Y + s.Cache.Static.CmdLineOutputH},
			rl.Vector2{X: s.RH / 2, Y: s.RH},
			s.P.Cur.Bg,
		)

		cp := rl.GetCharPressed()
		for ; cp > 0; cp = rl.GetCharPressed() {
			s.CommandLine.Prompt += string(rune(cp))
		}

		if rl.IsKeyPressed(rl.KeyEnter) {
			s.CommandLine.Lines = nil
			s.Cache.Static.FooterResizeF = true
			if len(s.CommandLine.Prompt) > 1 {
				s.WTX <- CommandPromptT{
					Prompt: strings.TrimPrefix(s.CommandLine.Prompt, ":"),
				}
			} else {
				s.CommandLine.Prompt = ""
			}
			s.M = Normal
		}

		if rl.IsKeyPressed(rl.KeyEscape) {
			s.CommandLine.Lines = nil
			s.Cache.Static.FooterResizeF = true
			s.CommandLine.Prompt = ""
			s.M = Normal
		}
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
			rl.Vector2{X: cl.p.X, Y: cl.p.Y + s.Cache.Static.CmdLineOutputH},
			s.CommandLine.Color,
		)
	}

	if s.M != Input && s.E == InputModeE {
		s.M = Input
		s.CommandLine.Prompt = ":"
		s.CommandLine.Color = s.P.Fg[1]
	}

	// cl.Outline(1, s.P.Base.Pink)
}

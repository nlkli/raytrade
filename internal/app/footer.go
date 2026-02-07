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
	if s.WRF {
		f.MoveTo(f.parent.p.X, f.parent.s.Y-FOOTER_H+RPD)
		f.SetSize(f.parent.s.X, FOOTER_H)
	}
	f.sl.Render(s)
	f.cl.Render(s)
}

type StatusLine struct {
	*Rect
	parent *Rect

	utS  string
	utTW int32
}

func (sl *StatusLine) Render(s *State) {
	if s.WRF {
		sl.MoveTo(sl.parent.p.X, sl.parent.p.Y)
		sl.SetSize(sl.parent.s.X, RH)
	}

	sl.Fill(s.P.Bg[0])

	if s.FN%uint64(s.TFPS/4) == 0 {
		sl.utS = time.Since(s.ST).Truncate(time.Second).String()
		sl.utTW = rl.MeasureText(sl.utS, RH_I32)
	}

	if len(sl.utS) > 0 {
		rl.DrawText(sl.utS, int32(sl.s.X)-sl.utTW, int32(sl.p.Y), RH_I32, s.P.Fg[3])
	}

	rl.DrawText(
		fmt.Sprintf("%s/%s", s.StatusLine.Symbol, s.StatusLine.Interval.AsString()),
		int32(sl.p.X), int32(sl.p.Y), RH_I32, s.P.Fg[3],
	)
}

type CommandLine struct {
	*Rect
	parent *Rect
}

func (cl *CommandLine) Render(s *State) {
	if s.WRF {
		cl.MoveTo(cl.parent.p.X, cl.parent.p.Y+RH)
		cl.SetSize(cl.parent.s.X, CLH)
	}

	if s.M == Input {
		if rl.IsKeyPressed(rl.KeyBackspace) {
			if len(s.CommandLine.Prompt) > 1 {
				r := []rune(s.CommandLine.Prompt)
				s.CommandLine.Prompt = string(r[:len(r)-1])
			}
		}

		tw := rl.MeasureText(s.CommandLine.Prompt, RH_I32)
		rl.DrawRectangle(int32(cl.p.X)+tw+2, int32(cl.p.Y), 8, RH_I32, s.P.Cur.Bg)

		cp := rl.GetCharPressed()
		for ; cp > 0; cp = rl.GetCharPressed() {
			s.CommandLine.Prompt += string(rune(cp))
		}

		if rl.IsKeyPressed(rl.KeyEnter) {
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
			s.CommandLine.Prompt = ""
			s.M = Normal
		}
	}

	if len(s.CommandLine.Prompt) > 0 {
		rl.DrawText(s.CommandLine.Prompt, int32(cl.p.X), int32(cl.p.Y), RH_I32, s.CommandLine.Color)
	}

	if s.M == Normal && s.E == InputModeE {
		s.M = Input
		s.CommandLine.Prompt = ":"
		s.CommandLine.Color = s.P.Fg[1]
	}
}

package app2

import (
	"strings"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type Footer struct {
	*Rect
	parent *Rect

	sl *StatusLine
	cl *CommandLine
}

func (f *Footer) Render(s *State) {
	if s.IsWindowResized() {
		f.MoveTo(f.parent.p.X, f.parent.s.Y-FOOTER_H+ROOT_PD)
		f.Resize(f.parent.s.X, FOOTER_H)
	}
	f.sl.Render(s)
	f.cl.Render(s)
}

type StatusLine struct {
	*Rect
	parent *Rect
}

func (sl *StatusLine) Render(s *State) {
	if s.IsWindowResized() {
		sl.MoveTo(sl.parent.p.X, sl.parent.p.Y)
		sl.SetSize(sl.parent.s.X, RH)
	}
	sl.Fill(s.P.Bg[0])
}

type CommandLine struct {
	*Rect
	parent *Rect
}

func (cl *CommandLine) Render(s *State) {
	if s.IsWindowResized() {
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
		tw := rl.MeasureText(s.CommandLine.Prompt, int32(RH))
		rl.DrawRectangle(int32(cl.p.X)+tw+2, int32(cl.p.Y), 8, int32(RH), s.P.Cur.Bg)
		if s.CPF {
			s.CommandLine.Prompt += string(s.CP)
			c := rl.GetCharPressed()
			for ; c > 0; c = rl.GetCharPressed() {
				s.CommandLine.Prompt += string(rune(c))
			}
		}
		if rl.IsKeyPressed(rl.KeyEnter) {
			if len(s.CommandLine.Prompt) > 1 {
				s.WTX <- CommandPromptTask{
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
		rl.DrawText(s.CommandLine.Prompt, int32(cl.p.X), int32(cl.p.Y), int32(RH), s.CommandLine.Color)
	}
	if s.M == Normal && s.CP == ':' {
		s.M = Input
		s.CommandLine.Prompt = ":"
		s.CommandLine.Color = s.P.Fg[1]
	}
}

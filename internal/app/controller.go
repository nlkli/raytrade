package app

import (
	"strings"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type Event int

const (
	NoneE Event = iota
	InputModeE
)

type Controller struct {
	CharLastPress struct {
		C  rune   // Char
		FN uint64 // Frame number
	}
}

func NewController() *Controller {
	return &Controller{}
}

func (c *Controller) Event(s *State) {

	if s.M == Input {
		if rl.IsKeyPressed(rl.KeyBackspace) && len(s.CommandLine.Prompt) > 1 {
			r := []rune(s.CommandLine.Prompt)
			s.CommandLine.Prompt = string(r[:len(r)-1])
		}

		cp := rl.GetCharPressed()
		for ; cp > 0; cp = rl.GetCharPressed() {
			s.CommandLine.Prompt += string(rune(cp))
			s.CommandLine.PromptW = s.StdMeasureText(s.CommandLine.Prompt).X
		}

		if rl.IsKeyPressed(rl.KeyEnter) {
			s.CommandLine.Lines = nil
			s.Footer.Forced = true

			if len(s.CommandLine.Prompt) > 1 {
				s.WTX <- CommandPromptT{
					Prompt: strings.TrimPrefix(s.CommandLine.Prompt, ":"),
				}
			} else {
				s.CommandLine.Prompt = ""
				s.CommandLine.PromptW = 0
			}

			s.M = Normal
		}

		if rl.IsKeyPressed(rl.KeyEscape) {
			s.CommandLine.Lines = nil
			s.Footer.Forced = true

			s.CommandLine.Prompt = ""
			s.CommandLine.PromptW = 0

			s.M = Normal
		}

		return
	}

	cp := rl.GetCharPressed()
	c.CharLastPress.C = cp
	c.CharLastPress.FN = s.FN

	if cp == 0 {
		return
	}

	if cp == ':' {
		s.M = Input
		s.CommandLine.Prompt = ":"
		s.CommandLine.PromptW = s.StdMeasureText(s.CommandLine.Prompt).X
		s.CommandLine.Color = s.P.Fg[1]
		return
	}
}

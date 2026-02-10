package app

import (
	"strings"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type Event int

const (
	NoneE Event = iota
	InputModeE
	CancelE
	SpaceE
	LastE
)

type Controller struct {
	LastEvent     Event
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
			s.CommandLine.PromptW = s.StdMeasureText(s.CommandLine.Prompt).X
		}

		cp := rl.GetCharPressed()
		for ; cp > 0; cp = rl.GetCharPressed() {
			s.CommandLine.Prompt += string(rune(cp))
			s.CommandLine.PromptW = s.StdMeasureText(s.CommandLine.Prompt).X
		}

		if rl.IsKeyPressed(rl.KeyEnter) {
			if len(s.CommandLine.Prompt) > 1 {
				s.CMDTX <- strings.TrimPrefix(s.CommandLine.Prompt, ":")

				if len(s.CommandLine.History) < COMMAND_LINE_HISTORY_CAP {
					s.CommandLine.History = append(s.CommandLine.History, s.CommandLine.Prompt)
				} else {
					copy(s.CommandLine.History, s.CommandLine.History[1:])
					s.CommandLine.History[len(s.CommandLine.History)-1] = s.CommandLine.Prompt
				}
			} else {
				s.CommandLine.Prompt = ""
				s.CommandLine.PromptW = 0
			}

			s.M = Normal
			s.CommandLine.Lines = nil
			s.Footer.Forced = true
			s.CommandLine.HustoryCur = -1
		}

		if rl.IsKeyPressed(rl.KeyUp) && len(s.CommandLine.History) > 0 {
			if s.CommandLine.HustoryCur == -1 {
				s.CommandLine.HustoryCur = len(s.CommandLine.History) - 1
			} else if s.CommandLine.HustoryCur > 0 {
				s.CommandLine.HustoryCur--
			}

			s.CommandLine.Prompt = s.CommandLine.History[s.CommandLine.HustoryCur]
			s.CommandLine.PromptW = s.StdMeasureText(s.CommandLine.Prompt).X
		}

		if rl.IsKeyPressed(rl.KeyDown) && s.CommandLine.HustoryCur != -1 {
			if s.CommandLine.HustoryCur < len(s.CommandLine.History)-1 {
				s.CommandLine.HustoryCur++
				s.CommandLine.Prompt = s.CommandLine.History[s.CommandLine.HustoryCur]
			} else {
				s.CommandLine.HustoryCur = -1
				s.CommandLine.Prompt = ":"
			}
			s.CommandLine.PromptW = s.StdMeasureText(s.CommandLine.Prompt).X
		}

		if rl.IsKeyPressed(rl.KeyEscape) {
			s.CommandLine.Lines = nil
			s.Footer.Forced = true

			s.CommandLine.Prompt = ""
			s.CommandLine.PromptW = 0
			s.CommandLine.HustoryCur = len(s.CommandLine.History)

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
		c.handleEvent(s, InputModeE)
		return
	}

	if rl.IsKeyPressed(rl.KeyEscape) {
		c.handleEvent(s, CancelE)
		return
	}
}

func (c *Controller) handleEvent(s *State, e Event) {
	c.LastEvent = e
	switch e {
	case InputModeE:
		s.M = Input
		s.CommandLine.Prompt = ":"
		s.CommandLine.PromptW = s.StdMeasureText(s.CommandLine.Prompt).X
		s.CommandLine.Color = s.P.Fg[1]
	default:
	}

}

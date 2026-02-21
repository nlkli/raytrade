package core

import (
	"strings"
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	REPEATED_PRESSING_DUR time.Duration = time.Millisecond * 555
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
	MouseHoldTime struct {
		LB time.Duration
		RB time.Duration
	}
	MousePressFrame struct {
		LB uint64
		RB uint64
	}
	CharLastPress struct {
		C  rune
		FN uint64
	}
	NumberBuf uint16
}

func NewController() *Controller {
	return &Controller{}
}

func (c *Controller) Event(s *State) {
	c.mouseEvent(s)

	if s.M == Input {
		c.handleInputMode(s)
		return
	}

	c.handleNormalMode(s)
}

func (c *Controller) mouseEvent(s *State) {
	const HOLDTIME = time.Duration(555) * time.Millisecond
	// const DOUBLEPRESS = time.Duration(333) * time.Microsecond

	if s.WHF || !s.WFF {
		s.Mouse.Captured = true
		return
	}

	s.Mouse.P = rl.GetMousePosition()
	s.Mouse.D = rl.GetMouseDelta()

	s.Mouse.PL = rl.IsMouseButtonPressed(rl.MouseButtonLeft)
	s.Mouse.DL = false
	if s.Mouse.PL {
		if s.FN-c.MousePressFrame.LB < 30 {
			s.Mouse.DL = true
		}
		c.MousePressFrame.LB = s.FN
	}

	s.Mouse.PR = rl.IsMouseButtonPressed(rl.MouseButtonRight)
	s.Mouse.DR = false
	if s.Mouse.PR {
		if s.FN-c.MousePressFrame.RB < 30 {
			s.Mouse.DR = true
		}
		c.MousePressFrame.RB = s.FN
	}

	if rl.IsMouseButtonDown(rl.MouseButtonLeft) {
		c.MouseHoldTime.LB += s.TFT
		if c.MouseHoldTime.LB > HOLDTIME {
			s.Mouse.HL = true
		}
	} else {
		s.Mouse.HL = false
		c.MouseHoldTime.LB = time.Duration(0)
	}

	if rl.IsMouseButtonDown(rl.MouseButtonRight) {
		c.MouseHoldTime.RB += s.TFT
		if c.MouseHoldTime.RB > HOLDTIME {
			s.Mouse.HR = true
		}
	} else {
		s.Mouse.HR = false
		c.MouseHoldTime.RB = time.Duration(0)
	}

	s.Mouse.Captured = false
}

func (c *Controller) handleInputMode(s *State) {
	// Backspace
	if rl.IsKeyPressed(rl.KeyBackspace) && len(s.CommandLine.Prompt) > 1 {
		r := []rune(s.CommandLine.Prompt)
		s.CommandLine.Prompt = string(r[:len(r)-1])

		s.CommandLine.PromptW = s.StdMeasureText(s.CommandLine.Prompt).X
	}

	// Char input
	for cp := rl.GetCharPressed(); cp > 0; cp = rl.GetCharPressed() {
		s.CommandLine.Prompt += string(rune(cp))

		s.CommandLine.PromptW = s.StdMeasureText(s.CommandLine.Prompt).X
	}

	// Enter
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
		s.CommandLine.HistoryCur = -1

		return
	}

	// History up
	if rl.IsKeyPressed(rl.KeyUp) && len(s.CommandLine.History) > 0 {

		if s.CommandLine.HistoryCur == -1 {
			s.CommandLine.HistoryCur = len(s.CommandLine.History) - 1
		} else if s.CommandLine.HistoryCur > 0 {
			s.CommandLine.HistoryCur--
		}

		s.CommandLine.Prompt = s.CommandLine.History[s.CommandLine.HistoryCur]

		s.CommandLine.PromptW = s.StdMeasureText(s.CommandLine.Prompt).X
	}

	// History down
	if rl.IsKeyPressed(rl.KeyDown) && s.CommandLine.HistoryCur != -1 {

		if s.CommandLine.HistoryCur < len(s.CommandLine.History)-1 {
			s.CommandLine.HistoryCur++
			s.CommandLine.Prompt = s.CommandLine.History[s.CommandLine.HistoryCur]
		} else {
			s.CommandLine.HistoryCur = -1
			s.CommandLine.Prompt = ":"
		}

		s.CommandLine.PromptW = s.StdMeasureText(s.CommandLine.Prompt).X
	}

	// Escape
	if rl.IsKeyPressed(rl.KeyEscape) {
		s.CommandLine.Lines = nil

		s.CommandLine.Prompt = ""
		s.CommandLine.PromptW = 0
		s.CommandLine.HistoryCur = len(s.CommandLine.History)

		s.M = Normal
	}
}

func (c *Controller) handleNormalMode(s *State) {
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

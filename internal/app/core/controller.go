package core

import (
	"strings"
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	REPEATED_PRESSING_DUR time.Duration = time.Millisecond * 555
)

type MouseEvent struct {
	Pos         rl.Vector2 // Position
	Delta       rl.Vector2 // Delta
	Click       [2]bool
	DoubleClick bool
	HoldLeft    bool
	Captured    bool
}

type Event struct {
	Mouse            MouseEvent
	ShiftDown        bool
	CtrlDown         bool
	SuperDown        bool
	CancelKeyPressed bool
}

type BindNode struct {
	Command  string
	Next     map[rune]*BindNode
	WithCtrl bool
}

type BindHit struct {
	Node *BindNode
	D    time.Duration
}

type Controller struct {
	CMDTX chan<- string // CMD tx

	Binds   map[rune]*BindNode
	BindHit *BindHit

	Mouse struct {
		HoldTime struct {
			LB time.Duration
			RB time.Duration
		}
		LastPressFrame struct {
			LB uint64
			RB uint64
		}
	}
}

func InitController(c *Config) *Controller {
	binds := make(map[rune]*BindNode, len(c.Binds))

	for _, b := range c.Binds {
		bindStr := b[0]
		command := b[1]

		if bindStr == "" {
			continue
		}

		bindStr = strings.ReplaceAll(bindStr, "<leader>", c.LeaderKey)

		currentMap := binds
		runes := []rune(bindStr)

		for i := 0; i < len(runes); i++ {
			r := runes[i]
			withCtrl := false

			// Ctrl combo
			if r == '<' && i+4 < len(runes) &&
				runes[i+1] == 'C' &&
				runes[i+2] == '-' &&
				runes[i+4] == '>' {

				r = runes[i+3]
				withCtrl = true
				i += 4
				continue
			}

			node, exists := currentMap[r]
			if !exists {
				node = &BindNode{
					Command: "",
					Next:    nil,
				}
			}

			node.WithCtrl = withCtrl

			if i == len(runes)-1 {
				node.Command = command
			}

			if i < len(runes)-1 && node.Next == nil {
				node.Next = make(map[rune]*BindNode)
			}

			currentMap[r] = node

			if node.Next != nil {
				currentMap = node.Next
			}
		}
	}

	return &Controller{
		Binds: binds,
	}
}

func (c *Controller) Event(s *State) {
	c.mouseEvent(s)

	if s.M == Input {
		c.handleInputMode(s)
		return
	}

	c.handleNormalMode(s)
}

func (c *Controller) handleNormalMode(s *State) {
	cp := rl.GetCharPressed()

	// c.LastCharPressed.C = cp
	// c.LastCharPressed.FN = s.FN

	s.E.ShiftDown = rl.IsKeyDown(rl.KeyLeftShift)
	s.E.CtrlDown = rl.IsKeyDown(rl.KeyLeftControl)
	s.E.SuperDown = rl.IsKeyDown(rl.KeyLeftSuper)

	if c.BindHit == nil {

		bind, ok := c.Binds[cp]

		if ok && bind.WithCtrl == s.E.CtrlDown && !s.E.SuperDown {

			if bind.Next == nil {
				c.CMDTX <- bind.Command
			} else {
				c.BindHit = &BindHit{
					Node: bind,
				}
			}

		}
	}

	if c.BindHit != nil {
		c.BindHit.D += s.ATFT

		bind, ok := c.BindHit.Node.Next[cp]

		if ok && bind.WithCtrl == s.E.CtrlDown && !s.E.SuperDown {

			if bind.Next == nil {
				c.CMDTX <- bind.Command
				c.BindHit = nil
			} else {
				c.BindHit = &BindHit{
					Node: bind,
				}
			}

		} else {

			if c.BindHit.D > time.Millisecond*777 {
				c.CMDTX <- c.BindHit.Node.Command
				c.BindHit = nil
			}

		}
	}

	if cp == 0 {
		return
	}

	if cp == ':' {
		s.M = Input
		s.CommandLine.Prompt = ":"
		s.CommandLine.PromptW = rl.MeasureTextEx(
			s.F,
			s.CommandLine.Prompt,
			s.RH,
			0,
		).X
		s.CommandLine.Color = s.P.Fg[1]

		c.handleInputMode(s)
		return
	}

	// if rl.IsKeyPressed(rl.KeyEscape) {
	// }
}

func (c *Controller) mouseEvent(s *State) {
	const HOLDTIME = time.Duration(333) * time.Millisecond

	if s.WHF || !s.WFF {
		s.E.Mouse.Captured = true
		return
	}

	s.E.Mouse.Pos = rl.GetMousePosition()
	s.E.Mouse.Delta = rl.Vector2Scale(rl.GetMouseDelta(), s.MouseDeltaFactor)

	s.E.Mouse.Click[0] = rl.IsMouseButtonReleased(rl.MouseButtonLeft)
	s.E.Mouse.Click[1] = rl.IsMouseButtonPressed(rl.MouseButtonRight)

	s.E.Mouse.DoubleClick = false
	if s.E.Mouse.Click[0] {
		if s.FN-c.Mouse.LastPressFrame.LB < uint64(s.TFPS)/2 {
			s.E.Mouse.DoubleClick = true
		}
		c.Mouse.LastPressFrame.LB = s.FN
	}

	if rl.IsMouseButtonDown(rl.MouseButtonLeft) {
		c.Mouse.HoldTime.LB += s.TFT
		if c.Mouse.HoldTime.LB > HOLDTIME {
			s.E.Mouse.HoldLeft = true
		}
	} else {
		s.E.Mouse.HoldLeft = false
		c.Mouse.HoldTime.LB = time.Duration(0)
	}

	s.E.Mouse.Captured = false
}

func (c *Controller) handleInputMode(s *State) {
	// Backspace
	if rl.IsKeyPressed(rl.KeyBackspace) && len(s.CommandLine.Prompt) > 1 {
		r := []rune(s.CommandLine.Prompt)
		s.CommandLine.Prompt = string(r[:len(r)-1])

		s.CommandLine.PromptW = rl.MeasureTextEx(
			s.F,
			s.CommandLine.Prompt,
			s.RH,
			0,
		).X
	}

	// Char input
	for cp := rl.GetCharPressed(); cp > 0; cp = rl.GetCharPressed() {
		s.CommandLine.Prompt += string(rune(cp))

		s.CommandLine.PromptW = rl.MeasureTextEx(
			s.F,
			s.CommandLine.Prompt,
			s.RH,
			0,
		).X
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

		s.CommandLine.PromptW = rl.MeasureTextEx(
			s.F,
			s.CommandLine.Prompt,
			s.RH,
			0,
		).X
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

		s.CommandLine.PromptW = rl.MeasureTextEx(
			s.F,
			s.CommandLine.Prompt,
			s.RH,
			0,
		).X
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

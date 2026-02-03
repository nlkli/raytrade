package components

import (
	"fmt"
	"nlkli/raytrade/internal/app/state"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	RH  float32 = 16        // Row height
	CLH float32 = RH * 2.33 // Command line height
)

type CommandLine struct {
	*Rect

	input string
}

func NewCommandLine() *CommandLine {
	return &CommandLine{
		Rect: NewRect(0, 0, 0, 0),
	}
}

func (cl *CommandLine) Render(s *state.State) state.Action {
	if s.IsWindowResized() {
		cl.MoveTo(0, s.W.Y-CLH)
		cl.Resize(s.W.X, CLH)
	}
	var action state.Action
	if s.PF {
		fmt.Println(s.CP)
	}
	if s.M != state.Input && s.CP == ':' {
		cl.input = ":"
		action = state.SetMode{Mode: state.Input}
	}
	if s.M == state.Input && len(cl.input) > 0 {
		rl.DrawText(cl.input, int32(cl.p.X), int32(cl.p.Y), int32(RH), s.P.Fg[1])
	}
	cl.Outline(1, s.P.Base.Red)
	return action
}

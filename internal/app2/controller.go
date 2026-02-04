package app2

import rl "github.com/gen2brain/raylib-go/raylib"

type Event int

const (
	EmptyE Event = iota
	InputModeE
)

type Controller struct {
}

func NewController() *Controller {
	return &Controller{}
}

func (c *Controller) Event(m Mode) Event {
	if m == Input {
		return EmptyE
	}

	cp := rl.GetCharPressed()
	if cp == 0 {
		return EmptyE
	}

	if cp == ':' {
		return InputModeE
	}

	return EmptyE
}

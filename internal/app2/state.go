package app2

import rl "github.com/gen2brain/raylib-go/raylib"

type Mode int

const (
	Normal Mode = iota
	Input
)

type State struct {
	FN uint64 // Frame number

	WR bool       // Window resized flag
	W  rl.Vector2 // Window size

	M Mode

	P *Palette

	CP rune // Char pressed
	PF bool // Char pressed flag

	CommandLinePrompt string
}

func InitState(c *Config) *State {
	return &State{
		W: rl.NewVector2(float32(c.InitWindow.Width), float32(c.InitWindow.Height)),
		M: Normal,
		P: PaletteFromConfig(c),
	}
}

func (s *State) IsWindowResized() bool {
	return s.FN == 0 || s.WR
}

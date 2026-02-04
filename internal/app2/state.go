package app2

import (
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type Mode int

const (
	Normal Mode = iota
	Input
)

type State struct {
	TFPS int32     // Target FPS
	ST   time.Time // Start time
	FN   uint64    // Frame number

	WRF bool // Window resized flag
	WHF bool // Window hidden flag
	WFF bool // Window focused flag

	WS rl.Vector2 // Window size

	M Mode

	P *Palette

	E Event // Controller event

	WTX chan<- Task // Worker tx

	StatusLine struct {
		Symbol   string
		Interval string
	}
	CommandLine struct {
		Prompt string
		Color  rl.Color
	}
}

func InitState(c *Config) *State {
	return &State{
		TFPS: c.TargetFPS,
		ST:   time.Now(),
		WS:   rl.NewVector2(float32(c.InitWindow.Width), float32(c.InitWindow.Height)),
		M:    Normal,
		P:    PaletteFromConfig(c),
	}
}

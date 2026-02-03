package state

import (
	"nlkli/raytrade/internal/app/config"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type Mode int

const (
	Normal Mode = iota
	Input
)

type State struct {
	FN uint64 // Frame number

	WR bool   // Window resized flag
	W rl.Vector2 // Window size

	M Mode

	P *Palette

	CP rune // Char pressed
	PF bool // Char pressed flag
}

func InitState(c *config.Config) *State {
	return &State{
		W: rl.NewVector2(float32(c.InitWin.Width), float32(c.InitWin.Height)),
		M: Normal,
		P: PaletteFromConfig(c),
	}
}

func (s *State) Update() {
	s.WR = rl.IsWindowResized()
	if s.WR {
		sW, sH := rl.GetScreenWidth(), rl.GetScreenHeight()
		s.W = rl.NewVector2(float32(sW), float32(sH))
	}
	cp := rl.GetCharPressed()
	if cp != 0 {
		s.PF = true
	} else {
		s.PF = false
	}
	s.CP = rune(cp)
}

func (s *State) IsWindowResized() bool {
	return s.FN == 0 || s.WR
}

func (s *State) Apply(a Action) {
	switch a := a.(type) {
	case SetMode:
		s.M = a.Mode
	}
}

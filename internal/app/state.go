package app

import (
	"nlkli/raytrade/internal/cdl"
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

	StatusLine  StatusLineState
	CommandLine CommandLineState
	Chart       ChatState
}

type StatusLineState struct {
	Symbol   string
	Interval cdl.Interval
}

type CommandLineState struct {
	Prompt string
	Color  rl.Color
}

type ChatState struct {
	shouldUpdate bool
	candleCh     chan cdl.CandleStreamData
	done         chan struct{}

	scale rl.Vector2
	shift rl.Vector2

	price      float64
	candles    []cdl.Candle
	winSize    int
	minP, maxP float64
	center     float64
	rng        float64
	offset     int
}

func InitState(c *Config) *State {
	return &State{
		TFPS: c.TargetFPS,
		ST:   time.Now(),
		WS:   rl.NewVector2(float32(c.InitWindow.Width), float32(c.InitWindow.Height)),
		M:    Normal,
		P:    PaletteFromConfig(c),
		Chart: ChatState{
			candleCh: make(chan cdl.CandleStreamData, 1),
			done:     make(chan struct{}, 1),

			scale: rl.NewVector2(1, 0.9),
			shift: rl.NewVector2(20, 0),
		},
	}
}

package core

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

const (
	COMMAND_LINE_HISTORY_CAP = 8
)

type CommitFn func(*State)

type State struct {
	ST time.Time // Start time
	FN uint64    // Frame number

	TFPS int32         // Target FPS
	TFT  time.Duration // Target frame time

	WRF bool // Window resized flag
	WHF bool // Window hidden flag
	WFF bool // Window focused flag

	WS rl.Vector2 // Window size

	ThrottlingF bool // frame num % (fps / 4) == 0

	M Mode

	P *Palette

	E     Event
	Mouse MouseState

	F        rl.Font
	RH       float32 // Row height
	RH_Dirty bool

	TextNumSV rl.Vector2 // Font base size number vector
	TextDotW  float32    // Font base size dot width

	BTX   chan<- Task   // Backgorund tx
	CMDTX chan<- string // CMD tx

	StatusLine  StatusLineState
	CommandLine CommandLineState

	Chart     []*ChartState
	OrderBook []*OrderBookState

	Bg BackgroundState

	ShowFPS     bool
	ShowOverlay bool

	Cache Cache
}

type MouseState struct {
	Pos         rl.Vector2 // Position
	Delta       rl.Vector2 // Delta
	Pressed     [2]bool
	DoubleClick bool
	HoldLeft    bool
	Captured    bool
}

type Cache struct {
	M      map[string]any
	Static struct {
	}
}

type BackgroundState struct {
}

type StatusLineState struct {
	Symbol   string
	Interval string
}

type CommandLineState struct {
	Prompt  string
	PromptW float32

	History    []string
	HistoryCur int // Cursor

	Lines  []string
	LinesH float32

	Color rl.Color
}

type ChartState struct {
	Forced bool // Forced update

	RHD float32

	Scale rl.Vector2 // X: candle scale, Y: price scale
	Shift rl.Vector2 // Pan offset (X: time, Y: price)

	Price  float64 // Last price
	PriceY float32 // Last price y coord

	Cursor      rl.Vector2
	CursorPrice float64

	Candles     []cdl.Candle // Candles buffer
	SecInterval float32      // Seconds interval

	Cap int // Number of candles that fit in canvas width

	MinP, MaxP float64 // Min/max price in visible range
	MidP       float64 // Average price
	RngP       float64 // Price range (MaxP - MinP)

	MaxVisPrice float64 // Topmost visible price value
	PxPerPrice  float64 // Conversion factor: pixels per one price unit

	StartSec float32 // Unix timestamp (sec) of first visible candle
	SecPerPx float32 // Conversion factor: seconds per one pixel

	ShowGrid  bool    // Toggle grid visibility
	GridStepY float32 // Pixels between horizontal grid lines (quantized)
	GridStepX float32 // Pixels between vertical grid lines (quantized)

	Levels []float64 // Price levels
}

type ChartLine struct {
	X     rl.Vector2 // X, Price
	Y     rl.Vector2 // X, Price
	Color rl.Color
}

type OrderBookState struct {
	Forced    bool // Forced update
	PlusCompI int

	Bids [][2]float64
	Asks [][2]float64
}

func InitState(c *Config) *State {
	palette := PaletteFromConfig(c)

	s := &State{
		TFPS: c.TargetFPS,
		TFT:  time.Second / time.Duration(c.TargetFPS),

		ST: time.Now(),
		WS: rl.NewVector2(
			float32(c.InitWindow.Width),
			float32(c.InitWindow.Height),
		),
		M: Normal,
		P: palette,

		RH: c.RowHeight,
		F:  rl.LoadFont(c.LoadFont),

		// StatusLine: StatusLineState{
		// 	Interval: cdl.M1.AsString(),
		// },

		CommandLine: CommandLineState{
			History: make([]string, 0, COMMAND_LINE_HISTORY_CAP),
			Color:   palette.Fg[1],
		},

		Cache: Cache{
			M: make(map[string]any),
		},
	}

	s.RH_Dirty = true

	const NUMBERS = "1234567890"

	s.TextNumSV = rl.MeasureTextEx(s.F, NUMBERS, float32(s.F.BaseSize), 0)
	s.TextNumSV.X = s.TextNumSV.X / float32(len(NUMBERS))

	s.TextDotW = rl.MeasureTextEx(
		s.F, ".", float32(s.F.BaseSize), 0,
	).X

	return s
}

func (s *State) StdDrawText(text string, pos rl.Vector2, color rl.Color) {
	rl.DrawTextEx(
		s.F,
		text,
		pos,
		s.RH,
		0,
		color,
	)
}

func (s *State) StdMeasureText(text string) rl.Vector2 {
	return rl.MeasureTextEx(s.F, text, s.RH, 0)
}

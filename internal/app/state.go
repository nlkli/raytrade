package app

import (
	"nlkli/raytrade/internal/cdl"
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	RPD float32 = 4 // Root padding

	OBW float32 = 200 // OrderBook section width

	TLH float32 = 20 // Time line height

	CW  float32 = 6.5 // Candle width
	CG  float32 = 2   // Candles gap
	CWW float32 = 1.5 // Candle wick width

	CMD_LINE_MARGIN_BOTTOM    float32 = 4
	TIME_LINE_LABELS_HEIGHT   float32 = 4
	PRICE_BAR_MAX_CONTENT     string  = ".0000000"
	PRICE_BAR_ROW_HEIGHT_DIFF float32 = -4

	PRICE_GRID_STEP_PX float64 = 40

	DEFAULT_ROW_HEIGHT float32 = 20
	DEFAULT_SCALE_X    float32 = 1
	DEFAULT_SCALE_Y    float32 = .9
	DEFAULT_SHIFT_X    float32 = 20
	DEFAULT_SHIFT_Y    float32 = 0
)

type Mode int

const (
	Normal Mode = iota
	Input
)

type State struct {
	TFPS   int32         // Target FPS
	TFPSFT time.Duration // Target FPS frame time

	ST time.Time // Start time
	FN uint64    // Frame number

	WRF bool // Window resized flag
	WHF bool // Window hidden flag
	WFF bool // Window focused flag

	WS rl.Vector2 // Window size

	M Mode

	P *Palette

	RH float32 // Row height
	F  rl.Font

	E Event // Controller event

	WTX chan<- Task // Worker tx

	StatusLine  StatusLineState
	CommandLine CommandLineState
	Chart       ChatState

	Cache Cache
}

type Cache struct {
	M      map[string]any
	Static struct {
		FooterH        float32
		FooterResizeF  bool
		CmdLineOutputH float32
		PriceBarW      float32
		TimeLineH      float32
	}
}

type StatusLineState struct {
	Symbol   string
	Interval cdl.Interval
}

type CommandLineState struct {
	Prompt string
	Lines  []string
	Color  rl.Color
}

type ChatState struct {
	shouldUpdate bool // forced update

	candleCh chan cdl.CandleStreamData
	done     chan struct{} // for close candle stream

	scale rl.Vector2
	shift rl.Vector2

	price  float64
	priceY float32 // price y coord

	candles    []cdl.Candle // candles buffer
	winSize    int          //
	minP, maxP float64      // min max price
	center     float64      // price center
	rng        float64      // price range

	yGrid [][2]float32 // [][y coord, price]

	// TODO ?
	offset int
}

func InitState(c *Config) *State {
	palette := PaletteFromConfig(c)
	return &State{
		TFPS:   c.TargetFPS,
		TFPSFT: time.Second / time.Duration(c.TargetFPS),
		ST:     time.Now(),
		WS:     rl.NewVector2(float32(c.InitWindow.Width), float32(c.InitWindow.Height)),
		M:      Normal,
		P:      palette,
		RH:     DEFAULT_ROW_HEIGHT,
		F:      rl.LoadFont(c.LoadFont),
		Chart: ChatState{
			shouldUpdate: true,

			candleCh: make(chan cdl.CandleStreamData, 1),
			done:     make(chan struct{}, 1),

			scale: rl.NewVector2(DEFAULT_SCALE_X, DEFAULT_SCALE_Y),
			shift: rl.NewVector2(DEFAULT_SHIFT_X, DEFAULT_SHIFT_Y),
		},
		CommandLine: CommandLineState{
			Color: palette.Fg[1],
		},
		Cache: Cache{
			M: make(map[string]any),
		},
	}
}

// Row height lvl
func (s *State) RHL(l int) float32 {
	const RHS float32 = 2
	switch l {
	case 0:
		return s.RH
	case 1:
		return max(RHS, s.RH-RHS)
	default:
		return max(RHS, s.RH-RHS*2)
	}
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

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

	COMMAND_LINE_HISTORY_CAP  int     = 8
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

type CommitFn func(*State) error

type State struct {
	ST time.Time // Start time
	FN uint64    // Frame number

	TFPS int32         // Target FPS
	TFT  time.Duration // Target FPS frame time
	// FT   time.Duration // Frame time

	WRF bool // Window resized flag
	WHF bool // Window hidden flag
	WFF bool // Window focused flag

	WS rl.Vector2 // Window size

	M Mode
	// E Event // Controller event

	P *Palette

	RH float32 // Row height
	F  rl.Font

	BTX   chan<- Task   // Backgorund tx
	CMDTX chan<- string // CMD tx

	Footer      FooterState
	StatusLine  StatusLineState
	CommandLine CommandLineState
	Chart       ChatState

	Cache Cache
}

type Cache struct {
	M      map[string]any
	Static struct {
	}
}

type FooterState struct {
	Height float32
	Forced bool
}

type StatusLineState struct {
	Symbol   string
	Interval string
}

type CommandLineState struct {
	History []string
	// TmpHistory []string
	HustoryCur int
	Prompt     string
	PromptW    float32
	Lines      []string
	LinesH     float32
	Color      rl.Color
}

type ChatState struct {
	Forced bool // forced update

	candleCh chan cdl.CandleStreamData
	done     chan struct{} // for close candle stream

	Scale rl.Vector2
	Shift rl.Vector2

	Price  float64
	PriceY float32 // price y coord

	Candles []cdl.Candle // candles buffer
	Cap     int          // canvas candles capacity

	MinP, MaxP float64 // min, max price
	CenterP    float64 // price center
	RangeP     float64 // price range

	ShowGrid bool

	GridY [][2]float32 // [][y coord, price]

	// TODO ?
	offset int

	PriceBarW float32
	TimeLineH float32
}

func InitState(c *Config) *State {
	palette := PaletteFromConfig(c)
	return &State{
		TFPS: c.TargetFPS,
		TFT:  time.Second / time.Duration(c.TargetFPS),

		ST: time.Now(),
		WS: rl.NewVector2(
			float32(c.InitWindow.Width),
			float32(c.InitWindow.Height),
		),
		M:  Normal,
		P:  palette,
		RH: DEFAULT_ROW_HEIGHT,
		F:  rl.LoadFont(c.LoadFont),
		Chart: ChatState{
			Forced: true,

			candleCh: make(chan cdl.CandleStreamData, 1),
			done:     make(chan struct{}, 1),

			Scale: rl.NewVector2(DEFAULT_SCALE_X, DEFAULT_SCALE_Y),
			Shift: rl.NewVector2(DEFAULT_SHIFT_X, DEFAULT_SHIFT_Y),

			ShowGrid: true,
		},
		CommandLine: CommandLineState{
			History: make([]string, 0, COMMAND_LINE_HISTORY_CAP),
			// TmpHistory: make([]string, COMMAND_LINE_HISTORY_CAP-1),
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

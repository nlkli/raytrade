package app

import (
	"nlkli/raytrade/internal/broker"
	"nlkli/raytrade/internal/cdl"
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	RPD float32 = 4 // Root padding

	OBW float32 = 0 // OrderBook section width

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
	DEFAULT_SHIFT_X    float32 = 40
	DEFAULT_SHIFT_Y    float32 = 0
)

type Mode int

const (
	Normal Mode = iota
	Input
)

type CommitFn func(*State)

func CommitCommandLineError(text string) CommitFn {
	return func(s *State) {
		s.CommandLine.Prompt = text
		s.CommandLine.Color = s.P.Base.Red
	}
}

func CommitCommandLineErrorAnd(text string, f CommitFn) CommitFn {
	return func(s *State) {
		CommitCommandLineError(text)(s)
		f(s)
	}
}

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

	Mouse MouseState

	RH float32 // Row height
	F  rl.Font

	BTX   chan<- Task   // Backgorund tx
	CMDTX chan<- string // CMD tx

	// Instrument InstrumentState

	Footer      FooterState
	StatusLine  StatusLineState
	CommandLine CommandLineState
	Chart       ChartState
	Bg          BackgroundState

	Cache Cache

	ShowFPS     bool
	ShowOverlay bool
}

type MouseState struct {
	P        rl.Vector2
	Captured bool
}

type Cache struct {
	M      map[string]any
	Static struct {
	}
}

// type InstrumentState struct {
// 	Category broker.Category
// 	Symbol   string
// 	Interval cdl.Interval
// }

type BackgroundState struct {
	IsActiveIO bool
	Category   broker.Category
	Symbol     string
	Interval   cdl.Interval
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
	HistoryCur int
	Prompt     string
	PromptW    float32
	Lines      []string
	LinesH     float32
	Color      rl.Color
}

type ChartState struct {
	Forced bool // Forced update

	Scale rl.Vector2
	Shift rl.Vector2

	Price  float64 // Last price
	PriceY float32 // Last price y coord

	CursorY     float32
	CursorPrice float64

	Candles []cdl.Candle // Candles buffer

	Cap        int     // Canvas candles capacity
	MinP, MaxP float64 // min, max price
	CenterP    float64 // Price center
	RangeP     float64 // Price range

	ShowGrid bool

	GridX [][2]float32 // [][x coord, seconds]
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
		// Instrument: InstrumentState{
		// 	Category: broker.Futures,
		// 	Interval: cdl.M5,
		// },
		StatusLine: StatusLineState{
			Interval: cdl.M1.AsString(),
		},
		Chart: ChartState{
			Forced: true,

			Scale: rl.NewVector2(DEFAULT_SCALE_X, DEFAULT_SCALE_Y),
			Shift: rl.NewVector2(DEFAULT_SHIFT_X, DEFAULT_SHIFT_Y),

			// GridX: make([][2]float32, ), // TODO

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

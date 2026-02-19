package core

import (
	"nlkli/raytrade/internal/broker"
	"nlkli/raytrade/internal/cdl"
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	RPD float32 = 4 // Root padding

	CW  float32 = 6.5 // Candle width
	CG  float32 = 2   // Candles gap
	CWW float32 = 1.5 // Candle wick width

	COMMAND_LINE_HISTORY_CAP int     = 8
	CMD_LINE_MARGIN_BOTTOM   float32 = 4

	TIME_LINE_RHL           int     = 2
	TIME_LINE_LABELS_HEIGHT float32 = 4

	PRICE_BAR_RHL             int     = 2
	PRICE_BAR_MAX_NUMBERS_CAP float32 = 7
	PRICE_BAR_MAX_CONTENT_CAP int     = 7 + 1 // Numbers + dot
	PRICE_BAR_LABLE_XPD       float32 = 4     // Padding

	ORDER_BOOK_WIDTH    float32 = 220
	ORDER_BOOK_RHL      int     = 1
	ORDER_BOOK_FILL_XPD float32 = 4 // Padding

	PRICE_GRID_STEP_PX float64 = 40

	DEFAULT_SCALE_X float32 = 1
	DEFAULT_SCALE_Y float32 = .9
	DEFAULT_SHIFT_X float32 = 40
	DEFAULT_SHIFT_Y float32 = 0
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
	TFT  time.Duration // Target frame time

	WRF bool // Window resized flag
	WHF bool // Window hidden flag
	WFF bool // Window focused flag

	WS rl.Vector2 // Window size

	M Mode

	P *Palette

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
	P        rl.Vector2
	Captured bool
}

type Cache struct {
	M      map[string]any
	Static struct {
	}
}

type BackgroundState struct {
	IsActiveIO bool
	Category   broker.Category
	Symbol     string
	Interval   cdl.Interval
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

	RHD int

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
}

type OrderBookState struct {
	Forced bool // Forced update

	RHD int

	Bids [][2]float64
	Asks [][2]float64

	Cap int

	MaxBidS,
	MaxAskS float64

	BidsText [][2]string
	AsksText [][2]string

	BidsPriceTextW []float32
	AsksPriceTextW []float32
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

		StatusLine: StatusLineState{
			Interval: cdl.M1.AsString(),
		},

		CommandLine: CommandLineState{
			History: make([]string, 0, COMMAND_LINE_HISTORY_CAP),
			Color:   palette.Fg[1],
		},

		Cache: Cache{
			M: make(map[string]any),
		},
	}

	s.SetRH(c.RowHeight)

	return s
}

func (s *State) RHL(l int) float32 {
	return max(2, s.RH-2*float32(l))
}

func (s *State) RHL_Scale(l int) float32 {
	return s.RHL(l) / float32(s.F.BaseSize)
}

func (s *State) SetRH(rh float32) {
	s.RH = max(2, rh)
	s.RH_Dirty = true

	const NUMBERS = "1234567890"

	s.TextNumSV = rl.MeasureTextEx(s.F, NUMBERS, float32(s.F.BaseSize), 0)
	s.TextNumSV.X = s.TextNumSV.X / float32(len(NUMBERS))

	s.TextDotW = rl.MeasureTextEx(
		s.F, ".", float32(s.F.BaseSize), 0,
	).X
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

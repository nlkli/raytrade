package core

import (
	"nlkli/raytrade/internal/broker"
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

	AFPS int32         // Actual FPS
	ATFT time.Duration // Actual frame time

	TFPS int32         // Target FPS
	TFT  time.Duration // Target frame time

	WRF bool // Window resized flag
	WHF bool // Window hidden flag
	WFF bool // Window focused flag

	WS rl.Vector2 // Window size

	ThrottlingF bool // frame num % (fps / 4) == 0

	M Mode

	P *Palette

	MouseDeltaFactor float32

	E Event

	F        rl.Font
	RH       float32 // Row height
	RH_Dirty bool

	TextNumSV rl.Vector2 // Font base size number vector
	TextDotW  float32    // Font base size dot width

	BTX   chan<- Task   // Backgorund tx
	CMDTX chan<- string // CMD tx

	InstrumentInfo map[string]InstrumentInfo

	Select Select

	StatusLine  StatusLineState
	CommandLine CommandLineState

	Chart     []*ChartState
	OrderBook []*OrderBookState

	Bg BackgroundState

	ShowOverlay bool

	Cache Cache
}

type InstrumentInfo struct {
	TickSize float64
}

type Select struct {
	OrderId struct {
		InstrumentKey string
		Value         float64
	}
	Price struct {
		InstrumentKey string
		Value         float64
	}
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

	InstrumentKey string
	Category      broker.Category
	Symbol        string
	Interval      cdl.Interval
	LableString   string

	Scale rl.Vector2 // X: candle scale, Y: price scale
	Shift rl.Vector2 // Pan offset (X: time, Y: price)

	ShowLable bool

	LastPrice  float64 // Last price
	LastPriceY float32 // Last price y coord

	Cursor      rl.Vector2
	CursorPrice float64

	ExtendCandlesF bool
	Candles        []cdl.Candle // Candles buffer
	SecInterval    float32      // Seconds interval

	CW  float32 // Candles width
	CWW float32 // Wick width
	CG  float32 // Candles gap

	Cap float32 // Number of candles that fit in canvas width

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

	Lines        [][2]rl.Vector2
	IsLineDuring bool
}

type OrderBookState struct {
	Forced    bool // Forced update
	PlusCompI int

	RHD float32

	ShiftY  float32 // for scroll centered
	OffsetY float32 // for scroll split

	InstrumentKey string
	Category      broker.Category
	Symbol        string

	Bids [][2]float64
	Asks [][2]float64
}

func (s *State) ApplyNewRH(rh float32) {
	const NUMBERS = "1234567890"

	s.RH = max(2, rh)
	s.RH_Dirty = true

	s.TextNumSV = rl.MeasureTextEx(
		s.F,
		NUMBERS,
		float32(s.F.BaseSize),
		0,
	)
	s.TextNumSV.X = s.TextNumSV.X / float32(len(NUMBERS))
	s.TextDotW = rl.MeasureTextEx(
		s.F, ".", float32(s.F.BaseSize), 0,
	).X
}

func (s *State) ApplyConfig(c *Config) {

	palette := PaletteFromConfig(c)
	targetFrameTime := time.Second / time.Duration(c.TargetFPS)

	rl.SetTargetFPS(c.TargetFPS)

	s.AFPS = c.TargetFPS
	s.ATFT = targetFrameTime

	s.TFPS = c.TargetFPS
	s.TFT = targetFrameTime

	s.P = palette
	s.F = rl.LoadFont(c.LoadFont)

	s.MouseDeltaFactor = c.MouseDeltaFactor
	s.CommandLine.Prompt = "configuration applied successfully"
	s.CommandLine.Color = palette.Fg[1]

	s.ApplyNewRH(c.RowHeight)

	s.ShowOverlay = c.ShowOverlay
}

func InitState(c *Config) *State {

	s := &State{
		ST: time.Now(),
		WS: rl.NewVector2(
			float32(c.InitWindow.Width),
			float32(c.InitWindow.Height),
		),
		M: Normal,

		CommandLine: CommandLineState{
			History: make([]string, 0, COMMAND_LINE_HISTORY_CAP),
		},

		Cache: Cache{
			M: make(map[string]any),
		},
	}

	s.ApplyConfig(c)

	return s
}

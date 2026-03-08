package comps

import (
	"fmt"
	"nlkli/raytrade/internal/app/core"
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	ROOT_PADDING float32 = 4

	DEFAULT_ORDERBOOK_VIEW_MODEL = 0

	DEFAULT_CHART_RHD     float32 = 4
	DEFAULT_ORDERBOOK_RHD float32 = 4

	DEFAULT_POSITION_RHD float32 = 2
	DEFAULT_ORDER_RHD    float32 = 2

	DEFAULT_CANDLE_WIDTH      float32 = 6.5
	DEFAULT_CANDLE_WICK_WIDTH float32 = 1.5
	DEFAULT_CANDLE_GAP        float32 = 2

	DEFAULT_SCALE_X float32 = .7
	DEFAULT_SCALE_Y float32 = .8
	DEFAULT_SHIFT_X float32 = 60
	DEFAULT_SHIFT_Y float32 = 0
)

// type CompType string
//
// const (
// 	SplitterCompType      CompType = "split"
// 	ChartCompType         CompType = "chart"
// 	OrderBookCompType     CompType = "orderbook"
// 	OrderBookPlusCompType CompType = "orderbook_plus"
// 	VoidCompType          CompType = "void"
// )

type Comp interface {
	R() *Rect
	Render(*core.State)
}

type Void struct {
	*Rect
}

func (v *Void) Render(s *core.State) {
}

func (v *Void) R() *Rect {
	return v.Rect
}

type Splitter struct {
	*Rect

	Axis int
	S    float32 // Size
	M    int     // Split mode

	A Comp
	B Comp
}

func (sp *Splitter) Render(s *core.State) {
	if s.WRF || s.RH_Dirty {
		as := sp.S
		if sp.Axis == 0 {
			switch sp.M {
			case 1:
				as = sp.s.Y - as
			case 2:
				as *= sp.s.Y
			}
			a, b := sp.SplitH(as)
			*sp.A.R() = *a
			*sp.B.R() = *b
		} else {
			switch sp.M {
			case 1:
				as = sp.s.X - as
			case 2:
				as *= sp.s.X
			}
			a, b := sp.SplitV(as)
			*sp.A.R() = *a
			*sp.B.R() = *b
		}
	}

	sp.A.Render(s)
	sp.B.Render(s)
}

func (s *Splitter) R() *Rect {
	return s.Rect
}

func getNumberParam[T ~float32 | ~float64 | ~int](params map[string]any, key string, defaultValue T) T {
	if val, ok := params[key]; ok {
		if f, ok := val.(float64); ok {
			return T(f)
		}
	}
	return defaultValue
}

func getBooleanParam(params map[string]any, key string, defaultValue bool) bool {
	if val, ok := params[key]; ok {
		if b, ok := val.(bool); ok {
			return b
		}
	}
	return defaultValue
}

func parseSplitter(c *core.Component, s *core.State) (Comp, error) {
	axis := getNumberParam(c.Params, "axis", 1)
	size := getNumberParam(c.Params, "size", .5)
	mode := getNumberParam(c.Params, "mode", 2)

	splitter := &Splitter{
		Rect: &Rect{},
		Axis: int(axis),
		S:    float32(size),
		M:    int(mode),
	}

	sA, err := parseComponentFromLayuotConfig(c.A, s)
	if err != nil {
		return nil, err
	}
	splitter.A = sA

	sB, err := parseComponentFromLayuotConfig(c.B, s)
	if err != nil {
		return nil, err
	}
	splitter.B = sB

	return splitter, nil
}

func parseChart(c *core.Component, s *core.State) (Comp, error) {
	chart := &Chart{
		Rect:     &Rect{},
		StateIdx: len(s.Chart),
	}

	chart.c.cam.Zoom = 1

	rhd := getNumberParam(c.Params, "rhd", DEFAULT_CHART_RHD)

	cw := getNumberParam(c.Params, "cw", DEFAULT_CANDLE_WIDTH)
	cww := getNumberParam(c.Params, "cww", DEFAULT_CANDLE_WICK_WIDTH)
	cg := getNumberParam(c.Params, "cg", DEFAULT_CANDLE_GAP)

	sx := getNumberParam(c.Params, "sx", DEFAULT_SCALE_X)
	sy := getNumberParam(c.Params, "sy", DEFAULT_SCALE_Y)
	tx := getNumberParam(c.Params, "tx", DEFAULT_SHIFT_X)
	ty := getNumberParam(c.Params, "ty", DEFAULT_SHIFT_Y)

	showLable := getBooleanParam(c.Params, "show_lable", true)
	showGrid := getBooleanParam(c.Params, "show_grid", true)
	showPosition := getBooleanParam(c.Params, "show_position", true)
	showOrder := getBooleanParam(c.Params, "show_order", true)
	showPriceBar := getBooleanParam(c.Params, "show_price_bar", true)
	showTimeLine := getBooleanParam(c.Params, "show_time_line", true)

	s.Chart = append(s.Chart, &core.ChartState{
		Forced: true,

		RHD: rhd,

		PositionIdx: -1,

		Scale: rl.Vector2{X: sx, Y: sy},
		Shift: rl.Vector2{X: tx, Y: ty},

		CW:  cw,
		CWW: cww,
		CG:  cg,

		ShowLable:    showLable,
		ShowGrid:     showGrid,
		ShowPosition: showPosition,
		ShowOrder:    showOrder,
		ShowPriceBar: showPriceBar,
		ShowTimeLine: showTimeLine,
	})

	return chart, nil
}

func parseOrderBook(c *core.Component, s *core.State) (Comp, error) {
	rhd := getNumberParam(c.Params, "rhd", DEFAULT_ORDERBOOK_RHD)
	vm := getNumberParam(c.Params, "vm", DEFAULT_ORDERBOOK_VIEW_MODEL)
	showText := getBooleanParam(c.Params, "show_text", true)

	orderBook := &OrderBook{
		Rect:     &Rect{},
		StateIdx: len(s.OrderBook),
		VM:       vm,
		ShowText: showText,
	}

	s.OrderBook = append(s.OrderBook, &core.OrderBookState{
		Forced: true,
		RHD:    rhd,
	})

	return orderBook, nil
}

func parseOrderBookPlus(c *core.Component, s *core.State) (Comp, error) {
	comp, err := parseSplitter(c, s)
	if err != nil {
		return nil, err
	}

	splitter := comp.(*Splitter)

	obA, ok := splitter.A.(*OrderBook)
	if !ok {
		return &Void{
			Rect: &Rect{},
		}, nil
	}

	obB, ok := splitter.B.(*OrderBook)
	if !ok {
		return &Void{
			Rect: &Rect{},
		}, nil
	}

	obB.StateIdx = obA.StateIdx
	s.OrderBook = s.OrderBook[:len(s.OrderBook)-1]
	s.OrderBook[len(s.OrderBook)-1].PlusCompI = 1

	orderBookPlus := &OrderBookPlus{
		splitter: splitter,
	}

	return orderBookPlus, nil
}

func parsePosition(c *core.Component, s *core.State) (Comp, error) {
	rhd := getNumberParam(c.Params, "rhd", DEFAULT_POSITION_RHD)

	s.Position.RHD = rhd
	s.Position.Forced = true

	return &Position{
		Rect: &Rect{},
	}, nil
}

func parseOrder(c *core.Component, s *core.State) (Comp, error) {
	rhd := getNumberParam(c.Params, "rhd", DEFAULT_ORDER_RHD)

	s.Order.RHD = rhd
	s.Order.Forced = true

	return &Order{
		Rect: &Rect{},
	}, nil
}

func parseComponentFromLayuotConfig(c *core.Component, s *core.State) (Comp, error) {
	switch c.Type {

	case "split":
		return parseSplitter(c, s)

	case "chart":
		return parseChart(c, s)

	case "orderbook":
		return parseOrderBook(c, s)

	case "orderbook_plus":
		return parseOrderBookPlus(c, s)

	case "position":
		return parsePosition(c, s)

	case "order":
		return parseOrder(c, s)

	default:
		return &Void{
			Rect: &Rect{},
		}, nil

	}
}

func InitRoot(c *core.Config, s *core.State) (*Root, error) {

	entryComp, err := parseComponentFromLayuotConfig(c.Layout, s)
	if err != nil {
		return nil, err
	}

	return &Root{
		Rect:      &Rect{},
		EntryComp: entryComp,
		Footer: Footer{
			Rect: &Rect{},
		},
	}, nil
}

type Root struct {
	*Rect

	EntryComp Comp
	Footer    Footer

	overlayTextV rl.Vector2
}

func (r *Root) Render(s *core.State) {
	rl.ClearBackground(s.P.Bg[1])

	if s.WRF || s.RH_Dirty {
		r.MoveTo(ROOT_PADDING, ROOT_PADDING)
		r.SetSize(s.WS.X-ROOT_PADDING*2, s.WS.Y-ROOT_PADDING*2)

		footerH := s.RH*2.2 + s.RH*float32(len(s.CommandLine.Lines))

		cr, fr := r.SplitH(r.s.Y - footerH)

		r.Footer.Rect = fr
		*r.EntryComp.R() = *cr
	}

	r.EntryComp.Render(s)
	r.Footer.Render(s)

	if s.ShowOverlay && r.s.X > 200 {
		overlayText := fmt.Sprintf(
			"StartTime: %s\nWindow: %v\nFrameNumber: %v\nFrameTime: %v\nFPS: %v",
			s.ST.Format(time.RFC1123),
			s.WS,
			s.FN,
			s.ATFT,
			s.AFPS,
		)
		if s.ThrottlingF {
			r.overlayTextV = rl.MeasureTextEx(
				s.F,
				overlayText,
				s.RH,
				0,
			)
		}
		rl.DrawRectangleV(
			rl.Vector2{X: ROOT_PADDING, Y: ROOT_PADDING},
			r.overlayTextV,
			s.P.OverlayBg,
		)
		rl.DrawTextEx(
			s.F,
			overlayText,
			rl.Vector2{X: ROOT_PADDING, Y: ROOT_PADDING},
			s.RH,
			0,
			s.P.Bright.Magenta,
		)
	}

	if s.RH_Dirty {
		s.RH_Dirty = false
	}
}

package comps

import (
	"errors"
	"fmt"
	"nlkli/raytrade/internal/app/core"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	ROOT_PADDING float32 = 4

	DEFAULT_CHART_RHD      float32 = 4
	DEFAULT_ORDER_BOOK_RHD float32 = 2

	DEFAULT_SCALE_X float32 = 1
	DEFAULT_SCALE_Y float32 = .9
	DEFAULT_SHIFT_X float32 = 40
	DEFAULT_SHIFT_Y float32 = 0
)

const ()

type Comp interface {
	R() *Rect
	Render(*core.State)
}

type Splitter struct {
	*Rect

	Axis int
	S    float32
	M    int

	A Comp
	B Comp
}

func (sp *Splitter) Render(s *core.State) {
	if s.WRF {
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

func parseSplitter(c *core.Component, s *core.State) (Comp, error) {
	axis := getNumberParam(c.Params, "axis", 1)
	size := getNumberParam(c.Params, "s", .5)
	mode := getNumberParam(c.Params, "m", 2)

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

func parseComponentFromLayuotConfig(c *core.Component, s *core.State) (Comp, error) {
	switch c.Type {
	case "split":
		return parseSplitter(c, s)
	case "chart":
		chart := &Chart{
			Rect: &Rect{},
			i:    len(s.Chart),
		}

		chart.c.cam.Zoom = 1

		rhd := getNumberParam(c.Params, "rhd", DEFAULT_CHART_RHD)
		sx := getNumberParam(c.Params, "sx", DEFAULT_SCALE_X)
		sy := getNumberParam(c.Params, "sy", DEFAULT_SCALE_Y)
		tx := getNumberParam(c.Params, "tx", DEFAULT_SHIFT_X)
		ty := getNumberParam(c.Params, "ty", DEFAULT_SHIFT_Y)

		showGrid := true
		if sgV, ok := c.Params["show_grid"]; ok {
			if sgB, ok := sgV.(bool); ok {
				showGrid = sgB
			}
		}

		s.Chart = append(s.Chart, &core.ChartState{
			Forced: true,

			RHD: rhd,

			Scale: rl.Vector2{X: sx, Y: sy},
			Shift: rl.Vector2{X: tx, Y: ty},

			ShowGrid: showGrid,
		})

		return chart, nil

	case "orderbook":

		rhd := getNumberParam(c.Params, "rhd", DEFAULT_ORDER_BOOK_RHD)
		vm := getNumberParam(c.Params, "vm", 0)

		showText := true
		if stV, ok := c.Params["show_text"]; ok {
			if stB, ok := stV.(bool); ok {
				showText = stB
			}
		}

		orderBook := &OrderBook{
			Rect:     &Rect{},
			I:        len(s.OrderBook),
			RHD:      rhd,
			VM:       vm,
			ShowText: showText,
		}

		s.OrderBook = append(s.OrderBook, &core.OrderBookState{
			Forced: true,
		})

		return orderBook, nil

	case "orderbook_plus":

		comp, err := parseSplitter(c, s)
		if err != nil {
			return nil, err
		}

		splitter := comp.(*Splitter)

		obA, ok := splitter.A.(*OrderBook)
		if !ok {
			return nil, errors.New("TODO")
		}

		obB, ok := splitter.B.(*OrderBook)
		if !ok {
			return nil, errors.New("TODO")
		}

		obB.I = obA.I
		s.OrderBook = s.OrderBook[:len(s.OrderBook)-1]
		s.OrderBook[len(s.OrderBook)-1].PlusCompI = 1

		orderBookPlus := &OrderBookPlus{
			splitter: splitter,
		}

		return orderBookPlus, nil

	default:
		return nil, errors.New("unknow component type")

	}
}

func InitRoot(c *core.Config, s *core.State) (*Root, error) {
	comp, err := parseComponentFromLayuotConfig(c.Layout, s)
	if err != nil {
		return nil, err
	}

	return &Root{
		Rect: &Rect{},
		c:    comp,
		f: Footer{
			Rect: &Rect{},
		},
	}, nil
}

type Root struct {
	*Rect

	c Comp
	f Footer
}

func (r *Root) Render(s *core.State) {
	rl.ClearBackground(s.P.Bg[1])

	if s.WRF || s.RH_Dirty {
		r.MoveTo(ROOT_PADDING, ROOT_PADDING)
		r.SetSize(s.WS.X-ROOT_PADDING*2, s.WS.Y-ROOT_PADDING*2)

		footerH := s.RH*2.2 + s.RH*float32(len(s.CommandLine.Lines))

		cr, fr := r.SplitH(r.s.Y - footerH)

		r.f.Rect = fr
		*r.c.R() = *cr
	}

	r.c.Render(s)
	r.f.Render(s)

	if s.ShowFPS {
		fps := fmt.Sprintf("%d", rl.GetFPS())
		s.StdDrawText(string(fps), rl.Vector2{X: ROOT_PADDING, Y: ROOT_PADDING}, s.P.Comment)
	}

	if s.RH_Dirty {
		s.RH_Dirty = false
	}
}

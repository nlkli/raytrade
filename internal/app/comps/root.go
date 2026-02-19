package comps

import (
	"errors"
	"fmt"
	"nlkli/raytrade/internal/app/core"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	ROOT_PADDING float32 = 4

	DEFAULT_CHART_RHD      int = 4
	DEFAULT_ORDER_BOOK_RHD int = 2

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

func parseComponentFromLayuotConfig(c *core.Component, s *core.State) (Comp, error) {
	switch c.Type {
	case "split":
		axisV, ok := c.Params["axis"]
		if !ok {
			return nil, errors.New("missing required field: axis")
		}
		axis, ok := axisV.(float64)
		if !ok {
			return nil, errors.New("field 'axis' must be a number")
		}

		sizeV, ok := c.Params["s"]
		if !ok {
			return nil, errors.New("missing required field: s")
		}
		size, ok := sizeV.(float64)
		if !ok {
			return nil, errors.New("field 's' must be a number")
		}

		modeV, ok := c.Params["m"]
		if !ok {
			return nil, errors.New("missing required field: m")
		}
		mode, ok := modeV.(float64)
		if !ok {
			return nil, errors.New("field 'm' must be a number")
		}

		if c.A == nil || c.B == nil {
			return nil, errors.New("split component requires both A and B children")
		}

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
	case "chart":
		chart := &Chart{
			Rect: &Rect{},
			i:    len(s.Chart),
		}

		chart.c.cam.Zoom = 1

		rhd := DEFAULT_CHART_RHD
		if rhdV, ok := c.Params["rhd"]; ok {
			if rhdF, ok := rhdV.(float64); ok {
				rhd = int(rhdF)
			}
		}

		sx := DEFAULT_SCALE_X
		if sxV, ok := c.Params["sx"]; ok {
			if sxF, ok := sxV.(float64); ok {
				sx = float32(sxF)
			}
		}

		sy := DEFAULT_SCALE_Y
		if syV, ok := c.Params["sy"]; ok {
			if syF, ok := syV.(float64); ok {
				sy = float32(syF)
			}
		}

		tx := DEFAULT_SHIFT_X
		if txV, ok := c.Params["tx"]; ok {
			if txF, ok := txV.(float64); ok {
				tx = float32(txF)
			}
		}

		ty := DEFAULT_SHIFT_Y
		if tyV, ok := c.Params["ty"]; ok {
			if tyF, ok := tyV.(float64); ok {
				ty = float32(tyF)
			}
		}

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
	case "order_book":
		orderBook := &OrderBook{
			Rect: &Rect{},
			i:    len(s.OrderBook),
		}

		rhd := DEFAULT_ORDER_BOOK_RHD
		if rhdV, ok := c.Params["rhd"]; ok {
			if rhdF, ok := rhdV.(float64); ok {
				rhd = int(rhdF)
			}
		}

		s.OrderBook = append(s.OrderBook, &core.OrderBookState{
			Forced: true,

			RHD: rhd,
		})

		return orderBook, nil
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

		footerH := s.RH*2.5 + s.RH*float32(len(s.CommandLine.Lines))

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

package comps

import (
	"errors"
	"fmt"
	"nlkli/raytrade/internal/app/core"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	ROOT_PADDING float32 = 4
)

const (
// Chart
)

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

func parseComponentFromLayuotConfig(c *core.Component) (Comp, error) {
	switch c.Type {
	case "split":
		axisVal, ok := c.Params["axis"]
		if !ok {
			return nil, errors.New("missing required field: axis")
		}
		axis, ok := axisVal.(float64)
		if !ok {
			return nil, errors.New("field 'axis' must be a number")
		}

		sVal, ok := c.Params["s"]
		if !ok {
			return nil, errors.New("missing required field: s")
		}
		s, ok := sVal.(float64)
		if !ok {
			return nil, errors.New("field 's' must be a number")
		}

		mVal, ok := c.Params["m"]
		if !ok {
			return nil, errors.New("missing required field: m")
		}
		m, ok := mVal.(float64)
		if !ok {
			return nil, errors.New("field 'm' must be a number")
		}

		if c.A == nil || c.B == nil {
			return nil, errors.New("split component requires both A and B children")
		}

		splitter := &Splitter{
			Rect: &Rect{},
			Axis: int(axis),
			S:    float32(s),
			M:    int(m),
		}

		sA, err := parseComponentFromLayuotConfig(c.A)
		if err != nil {
			return nil, err
		}
		splitter.A = sA

		sB, err := parseComponentFromLayuotConfig(c.B)
		if err != nil {
			return nil, err
		}
		splitter.B = sB

		return splitter, nil
	case "chart":
		return CreateChartComponent(c.Params), nil
	case "order_book":
		return CreateOrderBookComponent(c.Params), nil
	default:
		return nil, errors.New("unknow component type")

	}
}

func RootFromConfig(c *core.Config) (*Root, error) {
	comp, err := parseComponentFromLayuotConfig(c.Layout)
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

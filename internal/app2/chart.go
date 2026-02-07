package app2

import (
	"math"
	"nlkli/raytrade/internal/cdl"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type Chart struct {
	*Rect
	parent *Rect

	c  *Canvas
	tl *TimeLine
	pb *PriceBar
	cr *Crossing
}

func (ch *Chart) Render(s *State) {
	if s.WRF {
		ch.MoveTo(ch.parent.p.X, ch.parent.p.Y)
		ch.SetSize(ch.parent.s.X-OBW, ch.parent.s.Y)
	}

	ch.c.Render(s)
	ch.pb.Render(s)
	ch.tl.Render(s)
	ch.cr.Render(s)

	// ch.Outline(1, s.P.Base.Orange)
}

type Canvas struct {
	*Rect
	parent *Rect

	cam rl.Camera2D
}

func (c *Canvas) Render(s *State) {
	candles := s.Chart.candles
	n := len(candles)
	sf := s.Chart.scale

	if s.WRF {
		c.MoveTo(c.parent.p.X, c.parent.p.Y)
		c.SetSize(c.parent.s.X-PBW-RPD, c.parent.s.Y-TLH)

		c.cam.Offset = rl.NewVector2(c.p.X, c.p.Y)
		c.cam.Zoom = 1

		s.Chart.winSize = int((c.s.X + CW) / (CW*sf.X + CG))
		if n > 0 {
			s.Chart.minP, s.Chart.maxP = cdl.MinMaxPrice(
				candles[max(0, n-s.Chart.winSize-1):],
			)
		}
	}

	minP, maxP := s.Chart.minP, s.Chart.maxP
	if maxP == minP {
		return
	}

	center := (maxP + minP) * 0.5
	rangeY := (maxP - minP) / float64(sf.Y)
	w := CW * sf.X
	step := w + CG

	rl.BeginScissorMode(
		int32(c.p.X),
		int32(c.p.Y),
		int32(c.s.X),
		int32(c.s.Y),
	)

	c.cam.Target = s.Chart.shift
	rl.BeginMode2D(c.cam)

	for i := range candles {
		candle := candles[n-i-1]

		localX := c.s.X - step*float32(i+1)
		if localX+w < 0 {
			break
		}

		yO := float32((center-candle.O)/rangeY)*c.s.Y + c.s.Y/2
		yH := float32((center-candle.H)/rangeY)*c.s.Y + c.s.Y/2
		yL := float32((center-candle.L)/rangeY)*c.s.Y + c.s.Y/2
		yC := float32((center-candle.C)/rangeY)*c.s.Y + c.s.Y/2

		var color rl.Color
		if candle.C >= candle.O {
			color = s.P.Base.Green
		} else {
			color = s.P.Base.Red
		}

		DrawCandle(localX, yO, yH, yL, yC, w, color)
	}

	rl.EndMode2D()
	rl.EndScissorMode()
}

type TimeLine struct {
	*Rect
	parent *Rect
}

func (tl *TimeLine) Render(s *State) {
	if s.WRF {
		tl.MoveTo(tl.parent.p.X, tl.parent.s.Y-TLH+RPD)
		tl.SetSize(tl.parent.s.X-PBW-RPD, TLH)
	}

	tl.Fill(s.P.Bg[2])

	// tl.Outline(1, s.P.Base.Blue)
}

type PriceBar struct {
	*Rect
	parent *Rect
}

func (pb *PriceBar) Render(s *State) {
	if s.WRF {
		pb.MoveTo(pb.parent.s.X-PBW, pb.parent.p.Y)
		pb.SetSize(PBW+RPD, pb.parent.s.Y-TLH)
	}

	pb.Fill(s.P.Bg[0])

	// pb.Outline(1, s.P.Base.Green)
}

type Crossing struct {
	*Rect
	parent *Rect
}

func (cr *Crossing) Render(s *State) {
	if s.WRF {
		cr.MoveTo(cr.parent.s.X-PBW, cr.parent.s.Y-TLH+RPD)
		cr.SetSize(PBW+RPD, TLH)
	}

	cr.Fill(s.P.Bg[0])

	// cr.Outline(1, s.P.Base.Red)
}

func DrawCandle(x, yO, yH, yL, yC, width float32, color rl.Color) {
	halfW := x + width/2

	rl.DrawLineEx(
		rl.NewVector2(halfW, yH),
		rl.NewVector2(halfW, yL),
		CWW,
		color,
	)

	rl.DrawRectangleV(
		rl.NewVector2(x, min(yO, yC)),
		rl.NewVector2(width, float32(max(1, math.Abs(float64(yO-yC))))),
		color,
	)
}

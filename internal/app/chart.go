package app

import (
	"fmt"
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

	s.Chart.TimeLineH = s.RH + TIME_LINE_LABELS_HEIGHT
	if s.FN%uint64(s.TFPS) == 0 || s.Chart.Forced {
		s.Chart.PriceBarW = rl.MeasureTextEx(
			s.F,
			PRICE_BAR_MAX_CONTENT,
			s.RHL(2),
			0,
		).X
	}

	ch.c.Render(s)
	ch.pb.Render(s)
	ch.tl.Render(s)
	ch.cr.Render(s)

	if s.Chart.Forced {
		s.Chart.Forced = false
	}

	// ch.Outline(1, s.P.Base.Magenta)
}

type Canvas struct {
	*Rect
	parent *Rect

	cam rl.Camera2D
}

func (c *Canvas) Render(s *State) {
	candles := s.Chart.Candles
	n := len(candles)

	sf := s.Chart.Scale
	cw := CW * sf.X
	stepX := cw + CG

	if s.WRF || s.Chart.Forced {
		c.MoveTo(c.parent.p.X, c.parent.p.Y)
		c.SetSize(
			c.parent.s.X-c.parent.p.X-s.Chart.PriceBarW,
			c.parent.s.Y-s.Chart.TimeLineH,
		)

		c.cam.Offset = rl.Vector2{X: c.p.X, Y: c.p.Y}
		c.cam.Target = s.Chart.Shift
		c.cam.Zoom = 1

		if n == 0 {
			return
		}

		s.Chart.Cap = int((c.s.X - c.cam.Target.X + CW) / stepX)

		s.Chart.MinP, s.Chart.MaxP = cdl.MinMaxPrice(
			candles[max(0, n-s.Chart.Cap):],
		)
		s.Chart.CenterP = (s.Chart.MaxP + s.Chart.MinP) * .5
		s.Chart.RangeP = s.Chart.MaxP - s.Chart.MinP

		// Y grid

		pricePerPixel := s.Chart.RangeP / float64(sf.Y) / float64(c.s.Y)
		priceStep := pricePerPixel * float64(s.RH*2.2)

		qPriceStep := quantizePriceStep(priceStep)
		pixelStep := float32(qPriceStep / pricePerPixel)

		scaleY := c.s.Y / (float32(s.Chart.RangeP) / sf.Y)
		offsetY := c.s.Y * .5

		s.Chart.GridY = make([][2]float32, 0, int(c.s.Y/pixelStep)+1)
		for y := c.s.Y - pixelStep*.5; y > 0; y -= pixelStep {
			yPrice := float32(s.Chart.CenterP) - (y-offsetY)/scaleY
			s.Chart.GridY = append(s.Chart.GridY, [...]float32{y, yPrice})
		}
	}

	if n == 0 {
		return
	}

	center := s.Chart.CenterP
	scaleY := c.s.Y / (float32(s.Chart.RangeP) / sf.Y)
	offsetY := c.s.Y * .5

	s.Chart.Price = candles[n-1].C // TODO
	s.Chart.PriceY = float32(center-s.Chart.Price)*scaleY + offsetY - s.Chart.Shift.Y

	rl.BeginScissorMode(
		int32(c.p.X),
		int32(c.p.Y),
		int32(c.s.X),
		int32(c.s.Y),
	)
	rl.BeginMode2D(c.cam)

	rl.DrawLineEx(
		rl.Vector2{X: 0, Y: s.Chart.PriceY},
		rl.Vector2{X: c.s.X + c.cam.Target.X, Y: s.Chart.PriceY},
		1,
		s.P.Bg[4],
	)

	if s.Chart.ShowGrid {
		for _, g := range s.Chart.GridY {
			rl.DrawLineEx(
				rl.Vector2{X: 0, Y: g[0]},
				rl.Vector2{X: c.s.X + c.cam.Target.X, Y: g[0]},
				1,
				s.P.Bg[2],
			)
		}
	}

	for i := range n {
		candle := candles[n-i-1]

		localX := c.s.X - stepX*float32(i+1)
		if localX-stepX < c.p.X { // TODO
			break
		}

		yO := float32(center-candle.O)*scaleY + offsetY
		yH := float32(center-candle.H)*scaleY + offsetY
		yL := float32(center-candle.L)*scaleY + offsetY
		yC := float32(center-candle.C)*scaleY + offsetY

		color := s.P.Base.Green
		if candle.C < candle.O {
			color = s.P.Base.Red
		}

		// Draw candle

		wickX := localX + (cw * 0.5)
		rl.DrawLineEx(
			rl.Vector2{X: wickX, Y: yH},
			rl.Vector2{X: wickX, Y: yL},
			CWW,
			color,
		)

		bodyPosY := min(yO, yC)
		bodySizeY := yO - yC
		if bodySizeY < 0 {
			bodySizeY = -bodySizeY
		}
		if bodySizeY < 1 {
			bodySizeY = 1
		}

		rl.DrawRectangleV(
			rl.Vector2{X: localX, Y: bodyPosY},
			rl.Vector2{X: cw, Y: bodySizeY},
			color,
		)
	}

	rl.EndMode2D()
	rl.EndScissorMode()
}

type TimeLine struct {
	*Rect
	parent *Rect
}

func (tl *TimeLine) Render(s *State) {
	if s.WRF || s.Chart.Forced {
		tlh := s.Chart.TimeLineH

		tl.MoveTo(tl.parent.p.X, tl.parent.s.Y-tlh+tl.parent.p.Y)
		tl.SetSize(tl.parent.s.X-tl.parent.p.Y-s.Chart.PriceBarW, tlh)
	}

	// tl.Fill(s.P.Bg[2])

	// tl.Outline(1, s.P.Base.Blue)
}

type PriceBar struct {
	*Rect
	parent *Rect
}

func (pb *PriceBar) Render(s *State) {
	if s.WRF || s.Chart.Forced {
		tlh := s.Chart.TimeLineH
		pbw := s.Chart.PriceBarW

		pb.MoveTo(pb.parent.s.X-pbw, pb.parent.p.Y)
		pb.SetSize(pb.parent.p.X+pbw, pb.parent.s.Y-tlh)
	}

	if len(s.Chart.Candles) == 0 {
		return
	}

	localPriceY := s.Chart.PriceY + pb.p.Y

	rh := s.RHL(2) // Price bar row height
	rc := rh * .5  // Price bae row center

	rl.BeginScissorMode(
		int32(pb.p.X),
		int32(pb.p.Y),
		int32(pb.s.X),
		int32(pb.s.Y),
	)

	for _, g := range s.Chart.GridY {
		localTextY := pb.p.Y + g[0] - rc

		priceText := fmt.Sprintf("%.5f", g[1])
		if len(priceText) > len(PRICE_BAR_MAX_CONTENT) {
			priceText = priceText[:len(PRICE_BAR_MAX_CONTENT)]
		}

		rl.DrawTextEx(
			s.F,
			priceText,
			rl.Vector2{X: pb.p.X, Y: localTextY},
			rh,
			0,
			s.P.Fg[3],
		)
	}

	rl.DrawLineEx(
		rl.Vector2{X: pb.p.X, Y: localPriceY},
		rl.Vector2{X: pb.p.X + pb.s.X, Y: localPriceY},
		2,
		s.P.Base.Orange,
	)

	priceText := fmt.Sprintf("%.5f", s.Chart.Price)
	if len(priceText) > len(PRICE_BAR_MAX_CONTENT) {
		priceText = priceText[:len(PRICE_BAR_MAX_CONTENT)]
	}

	localTextY := localPriceY - rh

	rl.DrawTextEx(
		s.F,
		priceText,
		rl.Vector2{X: pb.p.X, Y: localTextY},
		rh,
		0,
		s.P.Fg[1],
	)

	// pb.Outline(1, s.P.Base.Red)

	rl.EndScissorMode()
}

type Crossing struct {
	*Rect
	parent *Rect
}

func (cr *Crossing) Render(s *State) {
	if s.WRF || s.Chart.Forced {
		tlh := s.Chart.TimeLineH
		pbw := s.Chart.PriceBarW

		cr.MoveTo(cr.parent.s.X-pbw, cr.parent.s.Y-tlh+cr.parent.p.Y)
		cr.SetSize(cr.parent.p.X+pbw, tlh)
	}

	// cr.Fill(s.P.Bg[0])

	// cr.Outline(1, s.P.Base.Red)
}

func quantizePriceStep(price float64) float64 {
	exp := math.Floor(math.Log10(price))
	base := math.Pow(10, exp)
	frac := price / base

	switch {
	case frac <= 1:
		return 1 * base
	case frac <= 2:
		return 2 * base
	case frac <= 2.5:
		return 2.5 * base
	case frac <= 5:
		return 5 * base
	default:
		return 10 * base
	}
}

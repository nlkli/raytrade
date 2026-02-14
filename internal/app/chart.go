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
		l, r := ch.SplitV(ch.s.X - s.Chart.PriceBarW)
		ch.c.Rect, ch.tl.Rect = l.SplitH(ch.s.Y - s.Chart.TimeLineH)
		ch.pb.Rect, ch.cr.Rect = r.SplitH(ch.s.Y - s.Chart.TimeLineH)
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

func priceToWorldY(price float64, maxVisible float64, scale float64) float32 {
	return float32((maxVisible - price) * scale)
}

func worldYToPrice(y float32, maxVisible float64, scale float64) float64 {
	return maxVisible - float64(y)/scale
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
	cw := CW * sf.X  // candle width * scale
	stepX := cw + CG // scale candle width + gap

	if s.WRF || s.Chart.Forced {
		c.cam.Offset = rl.Vector2{X: c.p.X, Y: c.p.Y}
		c.cam.Target = s.Chart.Shift
		c.cam.Zoom = 1

		if n == 0 {
			return
		}

		s.Chart.Cap = int(c.s.X / stepX)
		visibleC := candles[max(0, n-s.Chart.Cap):]

		s.Chart.MinP, s.Chart.MaxP = cdl.MinMaxPrice(visibleC)
		s.Chart.MidP = (s.Chart.MaxP + s.Chart.MinP) * .5
		s.Chart.RngP = s.Chart.MaxP - s.Chart.MinP

		// Y grid

		visiblePriceRange := s.Chart.RngP / float64(sf.Y)
		s.Chart.MaxVisiblePrice = s.Chart.MidP + visiblePriceRange*0.5

		s.Chart.PriceToPixel = float64(c.s.Y) / visiblePriceRange

		priceStep := quantizePriceStep(
			float64(s.Chart.PriceBarStep * float32(1/s.Chart.PriceToPixel)),
		)

		s.Chart.GridStepY = float32(priceStep * s.Chart.PriceToPixel)

		// s.Chart.SecInterval

	} else {

		if n == 0 {
			return
		}

		visiblePriceRange := s.Chart.RngP / float64(sf.Y)

		s.Chart.MaxVisiblePrice = s.Chart.MidP + visiblePriceRange*0.5

		s.Chart.PriceToPixel = float64(c.s.Y) / visiblePriceRange
	}

	s.Chart.Price = candles[n-1].C
	s.Chart.PriceY = priceToWorldY(
		s.Chart.Price, s.Chart.MaxVisiblePrice, s.Chart.PriceToPixel,
	)

	if s.Chart.ShowGrid {
		for y := c.s.Y - s.Chart.GridStepY*.5; y > 0; y -= s.Chart.GridStepY {
			localY := y + c.p.Y

			rl.DrawLineEx(
				rl.Vector2{X: c.p.X, Y: localY},
				rl.Vector2{X: c.p.X + c.s.X, Y: localY},
				1,
				s.P.Bg[2],
			)
		}
	}

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

	for i := range n {
		candle := candles[n-i-1]

		xp := c.s.X - stepX*float32(i+1)
		if xp-stepX < 0 { // TODO
			break
		}

		yO := priceToWorldY(candle.O, s.Chart.MaxVisiblePrice, s.Chart.PriceToPixel)
		yH := priceToWorldY(candle.H, s.Chart.MaxVisiblePrice, s.Chart.PriceToPixel)
		yL := priceToWorldY(candle.L, s.Chart.MaxVisiblePrice, s.Chart.PriceToPixel)
		yC := priceToWorldY(candle.C, s.Chart.MaxVisiblePrice, s.Chart.PriceToPixel)

		color := s.P.Base.Green
		if candle.C < candle.O {
			color = s.P.Base.Red
		}

		// Draw candle

		wickX := xp + (cw * 0.5)
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
			rl.Vector2{X: xp, Y: bodyPosY},
			rl.Vector2{X: cw, Y: bodySizeY},
			color,
		)
	}

	if !s.Mouse.Captured && c.ContainsV(s.Mouse.P) {
		s.Mouse.Captured = true

		worldMouse := rl.GetScreenToWorld2D(s.Mouse.P, c.cam)

		s.Chart.Cursor = worldMouse

		s.Chart.CursorPrice = worldYToPrice(
			s.Chart.Cursor.Y, s.Chart.MaxVisiblePrice, s.Chart.PriceToPixel,
		)

		rl.DrawLineEx(
			rl.Vector2{X: 0, Y: worldMouse.Y},
			rl.Vector2{X: c.s.X + c.cam.Target.X, Y: worldMouse.Y},
			1,
			s.P.Bg[4],
		)

		rl.DrawLineEx(
			rl.Vector2{X: worldMouse.X, Y: 0},
			rl.Vector2{X: worldMouse.X, Y: c.s.Y},
			1,
			s.P.Bg[4],
		)
	} else {
		s.Chart.Cursor = rl.Vector2{}
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
	}

	if len(s.Chart.Candles) == 0 {
		return
	}

	localPriceY := s.Chart.PriceY + pb.p.Y - s.Chart.Shift.Y

	rh := s.RHL(2) // Price bar row height
	rc := rh * .5  // Price bae row center

	pxInt32 := int32(pb.p.X)

	rl.BeginScissorMode(
		pxInt32,
		int32(pb.p.Y),
		int32(pb.s.X),
		int32(pb.s.Y),
	)

	for y := pb.s.Y - s.Chart.GridStepY*.5; y > 0; y -= s.Chart.GridStepY {
		localTextY := y + pb.p.Y - rc

		yPrice := worldYToPrice(
			y+s.Chart.Shift.Y, s.Chart.MaxVisiblePrice, s.Chart.PriceToPixel,
		)

		priceText := fmt.Sprintf("%.5f", yPrice)
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

	rl.DrawRectangleGradientV(
		pxInt32,
		int32(localPriceY-rh-4),
		int32(pb.s.X),
		int32(rh+4),
		rl.Color{},
		s.P.Bg[1],
	)

	rl.DrawRectangleGradientV(
		pxInt32,
		int32(localPriceY+2),
		int32(pb.s.X),
		int32(rh+4),
		s.P.Bg[1],
		rl.Color{},
	)

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

	if s.Chart.Cursor.Y > 0 {
		localCursorY := s.Chart.Cursor.Y + pb.p.Y - s.Chart.Shift.Y
		localTextY := localCursorY - rc

		priceText := fmt.Sprintf("%.5f", s.Chart.CursorPrice)
		if len(priceText) > len(PRICE_BAR_MAX_CONTENT) {
			priceText = priceText[:len(PRICE_BAR_MAX_CONTENT)]
		}

		rl.DrawRectangleGradientV(
			pxInt32,
			int32(localTextY-3),
			int32(pb.s.X),
			int32(rc+3),
			rl.Color{},
			s.P.Bg[1],
		)

		rl.DrawRectangleGradientV(
			pxInt32,
			int32(localCursorY),
			int32(pb.s.X),
			int32(rc+3),
			s.P.Bg[1],
			rl.Color{},
		)

		rl.DrawTextEx(
			s.F,
			priceText,
			rl.Vector2{X: pb.p.X, Y: localTextY},
			rh,
			0,
			s.P.Base.Blue,
		)
	}

	// pb.Outline(1, s.P.Base.Red)

	rl.EndScissorMode()
}

type Crossing struct {
	*Rect
	parent *Rect
}

func (cr *Crossing) Render(s *State) {
	if s.WRF || s.Chart.Forced {
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

func quantizeTimeStep(seconds float32) float32 {
	steps := []float32{
		60,      // 1m
		300,     // 5m
		900,     // 15m
		1800,    // 30m
		3600,    // 1h
		7200,    // 2h
		14400,   // 4h
		21600,   // 6h
		43200,   // 12h
		86400,   // 1d
		604800,  // 1w
		2592000, // 1m
	}

	for _, s := range steps {
		if s >= seconds {
			return s
		}
	}
	return steps[len(steps)-1]
}

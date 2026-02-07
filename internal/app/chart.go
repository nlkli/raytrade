package app

import (
	"fmt"
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
	w := CW * sf.X
	step := w + CG

	if s.WRF {
		c.MoveTo(c.parent.p.X, c.parent.p.Y)
		c.SetSize(c.parent.s.X-PBW-RPD, c.parent.s.Y-TLH)

		c.cam.Offset = rl.Vector2{X: c.p.X, Y: c.p.Y}
		c.cam.Zoom = 1
	}

	if s.WRF || s.Chart.shouldUpdate {
		s.Chart.winSize = int((c.s.X - c.cam.Target.X + CW) / step)
		if n > 0 {
			s.Chart.price = candles[n-1].C // TODO

			s.Chart.minP, s.Chart.maxP = cdl.MinMaxPrice(
				candles[max(0, n-s.Chart.winSize-1):],
			)
			s.Chart.center = (s.Chart.maxP + s.Chart.minP) * .5
			s.Chart.rng = (s.Chart.maxP - s.Chart.minP)
		}
		s.Chart.shouldUpdate = false
	}

	if n == 0 {
		return
	}

	center := s.Chart.center
	scaleY := c.s.Y / (float32(s.Chart.rng) / sf.Y)
	offsetY := c.s.Y * 0.5

	rl.BeginScissorMode(
		int32(c.p.X),
		int32(c.p.Y),
		int32(c.s.X),
		int32(c.s.Y),
	)

	c.cam.Target = s.Chart.shift
	rl.BeginMode2D(c.cam)

	for i := range n {
		candle := candles[n-i-1]

		localX := c.s.X - step*float32(i+1)
		if localX-step < c.p.X { // TODO
			break
		}

		yO := float32(center-candle.O)*scaleY + offsetY
		yH := float32(center-candle.H)*scaleY + offsetY
		yL := float32(center-candle.L)*scaleY + offsetY
		yC := float32(center-candle.C)*scaleY + offsetY

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

	// tl.Fill(s.P.Bg[2])

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

	center := s.Chart.center
	scaleY := pb.s.Y / (float32(s.Chart.rng) / s.Chart.scale.Y)
	offsetY := pb.s.Y * 0.5

	y := float32(center-s.Chart.price)*scaleY + offsetY - s.Chart.shift.Y
	screenY := pb.p.Y + y

	rl.DrawLineEx(
		rl.Vector2{X: pb.p.X, Y: screenY},
		rl.Vector2{X: pb.p.X + pb.s.X, Y: screenY},
		2,
		s.P.Base.Orange,
	)

	priceText := fmt.Sprintf("%.5f", s.Chart.price)
	if len(priceText) > 6 {
		priceText = priceText[:6]
	}
    // textY := scaleY
    
	rl.DrawText(
		priceText,
		int32(pb.p.X),
		int32(screenY)-PBRH_I32,
		PBRH_I32,
		s.P.Fg[1],
	)

	// pb.Fill(s.P.Bg[0])

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

	// cr.Fill(s.P.Bg[0])

	// cr.Outline(1, s.P.Base.Red)
}

func DrawCandle(x, yO, yH, yL, yC, width float32, color rl.Color) {
	halfW := x + width/2

	rl.DrawLineEx(
		rl.Vector2{X: halfW, Y: yH},
		rl.Vector2{X: halfW, Y: yL},
		CWW,
		color,
	)

	bodyPosY := min(yO, yC)
	bodySizeY := yO - yC
	if bodySizeY < 0 {
		bodySizeY *= -1
	}
	bodySizeY = max(1, bodySizeY)

	rl.DrawRectangleV(
		rl.Vector2{X: x, Y: bodyPosY},
		rl.Vector2{X: width, Y: bodySizeY},
		color,
	)
}

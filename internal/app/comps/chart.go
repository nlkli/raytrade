package comps

import (
	"fmt"
	"math"
	"nlkli/raytrade/internal/app/core"
	"nlkli/raytrade/internal/cdl"
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	CW  float32 = 6.5 // Candle width
	CG  float32 = 2   // Candles gap
	CWW float32 = 1.5 // Candle wick width

	TIME_LINE_RHL           int     = 2
	TIME_LINE_LABELS_HEIGHT float32 = 4

	PRICE_BAR_RHL             int     = 2
	PRICE_BAR_MAX_NUMBERS_CAP float32 = 7
	PRICE_BAR_MAX_CONTENT_CAP int     = 7 + 1 // Numbers + dot
	PRICE_BAR_LABLE_XPD       float32 = 4     // Padding
)

func CreateChartComponent(params map[string]any) Comp {
	return &Chart{
		Rect: &Rect{},
	}
}

type Chart struct {
	*Rect

	c  Canvas
	tl TimeLine
	pb PriceBar
	cr Crossing
}

func (ch *Chart) R() *Rect {
	return ch.Rect
}

func (ch *Chart) Render(s *core.State) {
	if s.WRF || s.RH_Dirty {
		priceBarW := (s.TextNumSV.X*
			PRICE_BAR_MAX_NUMBERS_CAP+s.TextDotW)*
			s.RHL_Scale(PRICE_BAR_RHL) + PRICE_BAR_LABLE_XPD*2
		timeLineH := s.RH + TIME_LINE_LABELS_HEIGHT

		l, r := ch.SplitV(ch.s.X - priceBarW)
		ch.c.Rect, ch.tl.Rect = l.SplitH(ch.s.Y - timeLineH)
		ch.pb.Rect, ch.cr.Rect = r.SplitH(ch.s.Y - timeLineH)
	}

	ch.c.Render(s)
	ch.pb.Render(s)
	ch.tl.Render(s)
	ch.cr.Render(s)

	// left gradient separator
	rl.DrawRectangleGradientH(
		int32(ch.p.X),
		int32(ch.p.Y),
		6,
		int32(ch.s.Y),
		s.P.Bg[1],
		s.P.TBg[1],
	)

	// top gradient separator
	rl.DrawRectangleGradientV(
		int32(ch.p.X),
		int32(ch.p.Y),
		int32(ch.s.X),
		6,
		s.P.Bg[1],
		s.P.TBg[1],
	)

	if s.Chart.Forced {
		s.Chart.Forced = false
	}
}

func priceToWorldY(price float64, maxVisible float64, scale float64) float32 {
	return float32((maxVisible - price) * scale)
}

func worldYToPrice(y float32, maxVisible float64, scale float64) float64 {
	return maxVisible - float64(y)/scale
}

type Canvas struct {
	*Rect

	cam rl.Camera2D
}

func (c *Canvas) Render(s *core.State) {
	cdls := s.Chart.Candles
	n := len(cdls)

	sf := s.Chart.Scale
	cw := CW * sf.X  // candle width * scale
	stepX := cw + CG // total width including gap

	if s.WRF || s.Chart.Forced || s.RH_Dirty {
		c.cam.Offset = rl.Vector2{X: c.p.X, Y: c.p.Y}
		c.cam.Target = s.Chart.Shift
		c.cam.Zoom = 1

		if n == 0 {
			return
		}

		// Visible candles count based on canvas width
		s.Chart.Cap = int(c.s.X / stepX)
		visC := cdls[max(0, n-s.Chart.Cap):]

		s.Chart.MinP, s.Chart.MaxP = cdl.MinMaxPrice(visC)
		s.Chart.MidP = (s.Chart.MaxP + s.Chart.MinP) * .5
		s.Chart.RngP = s.Chart.MaxP - s.Chart.MinP

		// Y-axis grid calculation
		visPriceRng := s.Chart.RngP / float64(sf.Y)
		s.Chart.MaxVisPrice = s.Chart.MidP + visPriceRng*0.5
		s.Chart.PxPerPrice = float64(c.s.Y) / visPriceRng

		priceStep := quantizePriceStep(
			float64(s.RH * 2.5 * float32(1/s.Chart.PxPerPrice)),
		)
		s.Chart.GridStepY = float32(priceStep * s.Chart.PxPerPrice)

		// X-axis grid calculation
		s.Chart.StartSec = float32(visC[0].Time / 1000)
		timeSecRng := float32(s.Chart.Cap) * s.Chart.SecInterval
		s.Chart.SecPerPx = timeSecRng / c.s.X

		timeStep := quantizeTimeStep(s.RH * 6 * s.Chart.SecPerPx)
		s.Chart.GridStepX = timeStep * (1 / s.Chart.SecPerPx)
	} else {
		if n == 0 {
			return
		}

		visPriceRng := s.Chart.RngP / float64(sf.Y)
		s.Chart.MaxVisPrice = s.Chart.MidP + visPriceRng*0.5
		s.Chart.PxPerPrice = float64(c.s.Y) / visPriceRng
	}

	// Current price position
	s.Chart.Price = cdls[n-1].C
	s.Chart.PriceY = priceToWorldY(
		s.Chart.Price, s.Chart.MaxVisPrice, s.Chart.PxPerPrice,
	)

	// Draw grid
	if s.Chart.ShowGrid {
		// Horizontal grid lines
		for y := c.s.Y - s.Chart.GridStepY*.5; y > 0; y -= s.Chart.GridStepY {
			localY := y + c.p.Y

			rl.DrawLineEx(
				rl.Vector2{X: c.p.X, Y: localY},
				rl.Vector2{X: c.p.X + c.s.X, Y: localY},
				1,
				s.P.Bg[2],
			)
		}
		// Vertical grid lines
		for x := c.s.X - s.Chart.GridStepX*.5; x > 0; x -= s.Chart.GridStepX {
			localX := x + c.p.X

			rl.DrawLineEx(
				rl.Vector2{X: localX, Y: c.p.Y},
				rl.Vector2{X: localX, Y: c.p.Y + c.s.Y},
				1,
				s.P.Bg[2],
			)
		}
	}

	// Clip to canvas area
	rl.BeginScissorMode(
		int32(c.p.X),
		int32(c.p.Y),
		int32(c.s.X),
		int32(c.s.Y),
	)
	rl.BeginMode2D(c.cam)

	// Current price line
	rl.DrawLineEx(
		rl.Vector2{X: 0, Y: s.Chart.PriceY},
		rl.Vector2{X: c.s.X + c.cam.Target.X, Y: s.Chart.PriceY},
		1,
		s.P.Bg[4],
	)

	// Draw candles from right to left (newest to oldest)
	for i := range n {
		candle := cdls[n-i-1]

		xPos := c.s.X - stepX*float32(i+1)
		if xPos-stepX < 0 {
			break
		}

		yO := priceToWorldY(candle.O, s.Chart.MaxVisPrice, s.Chart.PxPerPrice)
		yH := priceToWorldY(candle.H, s.Chart.MaxVisPrice, s.Chart.PxPerPrice)
		yL := priceToWorldY(candle.L, s.Chart.MaxVisPrice, s.Chart.PxPerPrice)
		yC := priceToWorldY(candle.C, s.Chart.MaxVisPrice, s.Chart.PxPerPrice)

		color := s.P.Base.Green
		if candle.C < candle.O {
			color = s.P.Base.Red
		}

		// Draw wick
		wickX := xPos + (cw * 0.5)
		rl.DrawLineEx(
			rl.Vector2{X: wickX, Y: yH},
			rl.Vector2{X: wickX, Y: yL},
			CWW,
			color,
		)

		// Draw body
		bodyY := min(yO, yC)
		bodyH := yO - yC
		if bodyH < 0 {
			bodyH = -bodyH
		}
		if bodyH < 1 {
			bodyH = 1 // Ensure minimum height for visibility
		}

		rl.DrawRectangleV(
			rl.Vector2{X: xPos, Y: bodyY},
			rl.Vector2{X: cw, Y: bodyH},
			color,
		)
	}

	// Mouse crosshair
	if !s.Mouse.Captured && c.ContainsV(s.Mouse.P) {
		s.Mouse.Captured = true

		worldMouse := rl.GetScreenToWorld2D(s.Mouse.P, c.cam)

		s.Chart.Cursor = worldMouse

		s.Chart.CursorPrice = worldYToPrice(
			s.Chart.Cursor.Y, s.Chart.MaxVisPrice, s.Chart.PxPerPrice,
		)

		// Horizontal line
		rl.DrawLineEx(
			rl.Vector2{X: 0, Y: worldMouse.Y},
			rl.Vector2{X: c.s.X + c.cam.Target.X, Y: worldMouse.Y},
			1,
			s.P.Bg[4],
		)

		// Vertical line
		rl.DrawLineEx(
			rl.Vector2{X: worldMouse.X, Y: 0},
			rl.Vector2{X: worldMouse.X, Y: c.s.Y + c.cam.Target.Y},
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
}

func (tl *TimeLine) Render(s *core.State) {
	if len(s.Chart.Candles) == 0 {
		return
	}

	rowH := s.RHL(TIME_LINE_RHL)
	// Text offset for centering
	halfText := (s.TextNumSV.X*2 + s.TextDotW/2) * s.RHL_Scale(TIME_LINE_RHL)

	rl.BeginScissorMode(
		int32(tl.p.X),
		int32(tl.p.Y),
		int32(tl.s.X),
		int32(tl.s.Y),
	)

	// Draw time labels at grid positions
	for x := tl.s.X - s.Chart.GridStepX*.5; x > 0; x -= s.Chart.GridStepX {
		localX := x + tl.p.X - halfText

		xSec := (x + s.Chart.Shift.X) * s.Chart.SecPerPx
		tUnix := time.Unix(int64(s.Chart.StartSec+xSec), 0)
		timeText := tUnix.Format("15:04")

		rl.DrawTextEx(
			s.F,
			timeText,
			rl.Vector2{X: localX, Y: tl.p.Y},
			rowH,
			0,
			s.P.Fg[3],
		)
	}

	rl.EndScissorMode()
}

type PriceBar struct {
	*Rect
}

func (pb *PriceBar) Render(s *core.State) {
	if len(s.Chart.Candles) == 0 {
		return
	}

	// Current price position in local coordinates
	localPriceY := s.Chart.PriceY + pb.p.Y - s.Chart.Shift.Y

	rowH := s.RHL(PRICE_BAR_RHL)
	rowCenter := rowH * .5
	// Lables offset x (padding)
	lable_xPos := pb.p.X + PRICE_BAR_LABLE_XPD

	rl.BeginScissorMode(
		int32(pb.p.X),
		int32(pb.p.Y),
		int32(pb.s.X),
		int32(pb.s.Y),
	)

	// Draw price labels at grid positions
	for y := pb.s.Y - s.Chart.GridStepY*.5; y > 0; y -= s.Chart.GridStepY {
		localTextY := y + pb.p.Y - rowCenter

		yPrice := worldYToPrice(
			y+s.Chart.Shift.Y, s.Chart.MaxVisPrice, s.Chart.PxPerPrice,
		)

		priceText := fmt.Sprintf("%.5f", yPrice)
		if len(priceText) > PRICE_BAR_MAX_CONTENT_CAP {
			priceText = priceText[:PRICE_BAR_MAX_CONTENT_CAP]
		}

		rl.DrawTextEx(
			s.F,
			priceText,
			rl.Vector2{X: lable_xPos, Y: localTextY},
			rowH,
			0,
			s.P.Fg[3],
		)
	}

	// Gradient highlights around current price line
	rl.DrawRectangleGradientV(
		int32(pb.p.X),
		int32(localPriceY-rowH-4),
		int32(pb.s.X),
		int32(rowH+4),
		s.P.TBg[1],
		s.P.Bg[1],
	)

	rl.DrawRectangleGradientV(
		int32(pb.p.X),
		int32(localPriceY+2),
		int32(pb.s.X),
		int32(rowH+4),
		s.P.Bg[1],
		s.P.TBg[1],
	)

	// Current price line
	rl.DrawLineEx(
		rl.Vector2{X: pb.p.X, Y: localPriceY},
		rl.Vector2{X: pb.p.X + pb.s.X, Y: localPriceY},
		2,
		s.P.Base.Orange,
	)

	// Current price label
	priceText := fmt.Sprintf("%.5f", s.Chart.Price)
	if len(priceText) > PRICE_BAR_MAX_CONTENT_CAP {
		priceText = priceText[:PRICE_BAR_MAX_CONTENT_CAP]
	}
	localTextY := localPriceY - rowH
	rl.DrawTextEx(
		s.F,
		priceText,
		rl.Vector2{X: lable_xPos, Y: localTextY},
		rowH,
		0,
		s.P.Fg[1],
	)

	// Cursor price indicator
	if s.Chart.Cursor.Y > 0 {
		localCursorY := s.Chart.Cursor.Y + pb.p.Y - s.Chart.Shift.Y
		localTextY := localCursorY - rowCenter

		priceText := fmt.Sprintf("%.5f", s.Chart.CursorPrice)
		if len(priceText) > PRICE_BAR_MAX_CONTENT_CAP {
			priceText = priceText[:PRICE_BAR_MAX_CONTENT_CAP]
		}

		// Gradient background for cursor price
		rl.DrawRectangleGradientV(
			int32(pb.p.X),
			int32(localTextY-3),
			int32(pb.s.X),
			int32(rowCenter+3),
			s.P.TBg[1],
			s.P.Bg[1],
		)

		rl.DrawRectangleGradientV(
			int32(pb.p.X),
			int32(localCursorY),
			int32(pb.s.X),
			int32(rowCenter+3),
			s.P.Bg[1],
			s.P.TBg[1],
		)

		rl.DrawTextEx(
			s.F,
			priceText,
			rl.Vector2{X: lable_xPos, Y: localTextY},
			rowH,
			0,
			s.P.Base.Blue,
		)
	}

	rl.EndScissorMode()
}

type Crossing struct {
	*Rect
}

func (cr *Crossing) Render(s *core.State) {
}

// quantizePriceStep rounds price to nice numbers for grid labels
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

// quantizeTimeStep rounds time to common intervals (seconds)
func quantizeTimeStep(sec float32) float32 {
	switch {
	case sec <= 60:
		return 60 // 1m
	case sec <= 300:
		return 300 // 5m
	case sec <= 900:
		return 900 // 15m
	case sec <= 1800:
		return 1800 // 30m
	case sec <= 3600:
		return 3600 // 1h
	case sec <= 7200:
		return 7200 // 2h
	case sec <= 14400:
		return 14400 // 4h
	case sec <= 21600:
		return 21600 // 6h
	case sec <= 43200:
		return 43200 // 12h
	case sec <= 86400:
		return 86400 // 1d
	case sec <= 604800:
		return 604800 // 1w
	default:
		return 2592000 // 30d
	}
}

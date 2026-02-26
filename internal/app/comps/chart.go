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

	PRICE_BAR_MAX_NUMBERS_CAP float32 = 7
	PRICE_BAR_MAX_CONTENT_CAP int     = 7 + 1 // Numbers + dot
	PRICE_BAR_LABLE_XPD       float32 = 4     // Padding
)

type Chart struct {
	*Rect

	ShowLable bool

	StateIdx int

	c  Canvas
	tl TimeLine
	pb PriceBar
	cr Crossing
}

func (ch *Chart) R() *Rect {
	return ch.Rect
}

func (ch *Chart) Render(s *core.State) {
	cs := s.Chart[ch.StateIdx] // Chart state
	rh := s.RH - cs.RHD

	if s.WRF || s.RH_Dirty {
		priceBarW := (s.TextNumSV.X*PRICE_BAR_MAX_NUMBERS_CAP+s.TextDotW)*
			(rh/float32(s.F.BaseSize)) +
			PRICE_BAR_LABLE_XPD*2

		l, r := ch.SplitV(ch.s.X - priceBarW)
		ch.c.Rect, ch.tl.Rect = l.SplitH(ch.s.Y - rh)
		ch.pb.Rect, ch.cr.Rect = r.SplitH(ch.s.Y - rh)
	}

	if len(cs.Candles) == 0 {
		return
	}

	ch.c.Render(s, cs)
	ch.pb.Render(s, cs, rh)
	ch.tl.Render(s, cs, rh)
	// ch.cr.Render(s, cs)

	// left gradient separator
	rl.DrawRectangleGradientH(
		int32(ch.p.X),
		int32(ch.p.Y),
		8,
		int32(ch.s.Y),
		s.P.Bg[1],
		s.P.TBg[1],
	)

	// top gradient separator
	rl.DrawRectangleGradientV(
		int32(ch.p.X),
		int32(ch.p.Y),
		int32(ch.s.X),
		8,
		s.P.Bg[1],
		s.P.TBg[1],
	)

	if ch.ShowLable && ch.s.X > 200 {
		rl.DrawTextEx(
			s.F,
			cs.LableString,
			rl.Vector2{
				X: ch.p.X + 2,
				Y: ch.p.Y + 2,
			},
			rh,
			0,
			s.P.Comment,
		)
	}

	if cs.Forced {
		cs.Forced = false
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

func (c *Canvas) Render(s *core.State, cs *core.ChartState) {
	candles := cs.Candles
	n := len(candles)

	sf := cs.Scale
	cw := CW * sf.X  // candle width * scale
	stepX := cw + CG // total width including gap

	c.cam.Offset = c.p
	c.cam.Target = cs.Shift

	if s.WRF || cs.Forced || s.RH_Dirty {
		// Visible candles count based on canvas width
		cs.Cap = c.s.X / stepX
		visC := candles[max(0, n-int(cs.Cap)):]

		cs.MinP, cs.MaxP = cdl.MinMaxPrice(visC)
		cs.MidP = (cs.MaxP + cs.MinP) * .5
		cs.RngP = cs.MaxP - cs.MinP

		// Y-axis grid calculation
		visPriceRng := cs.RngP / float64(sf.Y)
		cs.MaxVisPrice = cs.MidP + visPriceRng*0.5
		cs.PxPerPrice = float64(c.s.Y) / visPriceRng

		priceStep := quantizePriceStep(
			float64(s.RH * 2.5 * float32(1/cs.PxPerPrice)),
		)
		cs.GridStepY = float32(priceStep * cs.PxPerPrice)

		// X-axis grid calculation
		timeSecRng := cs.Cap * cs.SecInterval
		cs.StartSec = float32(visC[len(visC)-1].Time)/1000 - timeSecRng
		cs.SecPerPx = timeSecRng / c.s.X

		timeStep := quantizeTimeStep(s.RH * 6 * cs.SecPerPx)
		cs.GridStepX = timeStep * (1 / cs.SecPerPx)
	} else {
		visPriceRng := cs.RngP / float64(sf.Y)
		cs.MaxVisPrice = cs.MidP + visPriceRng*0.5
		cs.PxPerPrice = float64(c.s.Y) / visPriceRng
	}

	// Current price position
	cs.Price = candles[n-1].C
	cs.PriceY = priceToWorldY(
		cs.Price, cs.MaxVisPrice, cs.PxPerPrice,
	)

	// Draw grid
	if cs.ShowGrid {
		// Horizontal grid lines
		for y := c.s.Y - cs.GridStepY*.5; y > 0; y -= cs.GridStepY {
			localY := y + c.p.Y

			rl.DrawLineEx(
				rl.Vector2{X: c.p.X, Y: localY},
				rl.Vector2{X: c.p.X + c.s.X, Y: localY},
				1,
				s.P.Bg[2],
			)
		}
		// Vertical grid lines
		for x := c.s.X - cs.GridStepX*.5; x > 0; x -= cs.GridStepX {
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

	if len(cs.Levels) > 0 {
		for _, l := range cs.Levels {
			levelY := priceToWorldY(
				l, cs.MaxVisPrice, cs.PxPerPrice,
			)

			if levelY < 0 || levelY > c.s.Y {
				continue
			}

			rl.DrawLineEx(
				rl.Vector2{X: cs.Shift.X, Y: levelY},
				rl.Vector2{X: c.s.X + cs.Shift.X, Y: levelY},
				1,
				s.P.Base.Cyan,
			)
		}
	}

	if len(cs.Lines) > 0 {

		lines := cs.Lines

		if cs.IsLineDuring {
			worldMouse := rl.GetScreenToWorld2D(s.E.Mouse.Pos, c.cam)

			l := lines[len(lines)-1]

			lineX0 := (l[0].X - cs.StartSec) / cs.SecPerPx
			lineY0 := priceToWorldY(
				float64(l[0].Y), cs.MaxVisPrice, cs.PxPerPrice,
			)

			rl.DrawLineEx(
				rl.Vector2{X: lineX0, Y: lineY0},
				worldMouse,
				1,
				s.P.Dim.Blue,
			)

			lines = lines[:len(lines)-1]
		}

		for _, l := range lines {
			lineX0 := (l[0].X - cs.StartSec) / cs.SecPerPx
			lineX1 := (l[1].X - cs.StartSec) / cs.SecPerPx

			lineY0 := priceToWorldY(
				float64(l[0].Y), cs.MaxVisPrice, cs.PxPerPrice,
			)
			lineY1 := priceToWorldY(
				float64(l[1].Y), cs.MaxVisPrice, cs.PxPerPrice,
			)

			rl.DrawLineEx(
				rl.Vector2{X: lineX0, Y: lineY0},
				rl.Vector2{X: lineX1, Y: lineY1},
				1,
				s.P.Dim.Blue,
			)
		}
	}

	// Current price line
	rl.DrawLineEx(
		rl.Vector2{X: cs.Shift.X, Y: cs.PriceY},
		rl.Vector2{X: c.s.X + cs.Shift.X, Y: cs.PriceY},
		1,
		s.P.Bg[4],
	)

	start := 1
	if cs.Shift.X < 0 {
		start -= int(cs.Shift.X / stepX)
	}
	end := min(n, int(cs.Cap)+start+2)

	// Draw candles from right to left (newest to oldest)
	for i := start; i < end; i++ {
		candle := candles[n-i]

		xPos := c.s.X - stepX*float32(i)

		yO := priceToWorldY(candle.O, cs.MaxVisPrice, cs.PxPerPrice)
		yH := priceToWorldY(candle.H, cs.MaxVisPrice, cs.PxPerPrice)
		yL := priceToWorldY(candle.L, cs.MaxVisPrice, cs.PxPerPrice)
		yC := priceToWorldY(candle.C, cs.MaxVisPrice, cs.PxPerPrice)

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
	if !s.E.Mouse.Captured && c.ContainsV(s.E.Mouse.Pos) {

		wm := rl.GetScreenToWorld2D(s.E.Mouse.Pos, c.cam) // World mouse

		cs.Cursor = wm

		cs.CursorPrice = worldYToPrice(
			cs.Cursor.Y, cs.MaxVisPrice, cs.PxPerPrice,
		)

		// Horizontal line
		rl.DrawLineEx(
			rl.Vector2{X: cs.Shift.X, Y: wm.Y},
			rl.Vector2{X: c.s.X + cs.Shift.X, Y: wm.Y},
			1,
			s.P.Bg[4],
		)

		// Vertical line
		rl.DrawLineEx(
			rl.Vector2{X: wm.X, Y: cs.Shift.Y},
			rl.Vector2{X: wm.X, Y: c.s.Y + cs.Shift.Y},
			1,
			s.P.Bg[4],
		)

		if s.E.Mouse.HoldLeft {

			cs.Shift.X -= s.E.Mouse.Delta.X
			cs.Shift.Y -= s.E.Mouse.Delta.Y

		} else {

			if s.E.Mouse.Click[0] && cs.IsLineDuring {
				cs.Lines[len(cs.Lines)-1][1] = rl.Vector2{
					X: cs.StartSec + wm.X*cs.SecPerPx,
					Y: float32(
						worldYToPrice(wm.Y, cs.MaxVisPrice, cs.PxPerPrice),
					),
				}
				cs.IsLineDuring = false
			}

			if s.E.Mouse.DoubleClick && !cs.IsLineDuring {
				cs.Lines = append(
					cs.Lines,
					[2]rl.Vector2{
						{
							X: cs.StartSec + wm.X*cs.SecPerPx,
							Y: float32(
								worldYToPrice(wm.Y, cs.MaxVisPrice, cs.PxPerPrice),
							),
						},
						{},
					},
				)
				cs.IsLineDuring = true
			}

			if s.E.Mouse.Click[1] && cs.IsLineDuring {
				cs.Lines = cs.Lines[:len(cs.Lines)-1]
				cs.IsLineDuring = false
			}

		}

		s.E.Mouse.Captured = true
	} else {
		cs.Cursor = rl.Vector2{}
	}

	rl.EndMode2D()
	rl.EndScissorMode()
}

type TimeLine struct {
	*Rect
}

func (tl *TimeLine) Render(s *core.State, cs *core.ChartState, rh float32) {
	// Text offset for centering
	halfText := (s.TextNumSV.X*2 + s.TextDotW/2) * (rh / float32(s.F.BaseSize))

	rl.BeginScissorMode(
		int32(tl.p.X),
		int32(tl.p.Y),
		int32(tl.s.X),
		int32(tl.s.Y),
	)

	// Draw time labels at grid positions
	for x := tl.s.X - cs.GridStepX*.5; x > 0; x -= cs.GridStepX {
		localX := x + tl.p.X - halfText

		xSec := (x + cs.Shift.X) * cs.SecPerPx
		tUnix := time.Unix(int64(cs.StartSec+xSec), 0)
		timeText := tUnix.Format("15:04")

		rl.DrawTextEx(
			s.F,
			timeText,
			rl.Vector2{X: localX, Y: tl.p.Y},
			rh,
			0,
			s.P.Fg[3],
		)
	}

	rl.EndScissorMode()
}

type PriceBar struct {
	*Rect
}

func (pb *PriceBar) Render(s *core.State, cs *core.ChartState, rh float32) {
	// Current price position in local coordinates
	localPriceY := cs.PriceY + pb.p.Y - cs.Shift.Y

	rowCenter := rh * .5
	// Lables offset x (padding)
	lable_xPos := pb.p.X + PRICE_BAR_LABLE_XPD

	rl.BeginScissorMode(
		int32(pb.p.X),
		int32(pb.p.Y),
		int32(pb.s.X),
		int32(pb.s.Y),
	)

	// Draw price labels at grid positions
	for y := pb.s.Y - cs.GridStepY*.5; y > 0; y -= cs.GridStepY {
		localTextY := y + pb.p.Y - rowCenter

		yPrice := worldYToPrice(
			y+cs.Shift.Y, cs.MaxVisPrice, cs.PxPerPrice,
		)

		priceText := fmt.Sprintf("%.5f", yPrice)
		if len(priceText) > PRICE_BAR_MAX_CONTENT_CAP {
			priceText = priceText[:PRICE_BAR_MAX_CONTENT_CAP]
		}

		rl.DrawTextEx(
			s.F,
			priceText,
			rl.Vector2{X: lable_xPos, Y: localTextY},
			rh,
			0,
			s.P.Fg[3],
		)
	}

	// Gradient highlights around current price line
	rl.DrawRectangleGradientV(
		int32(pb.p.X),
		int32(localPriceY-rh-4),
		int32(pb.s.X),
		int32(rh+4),
		s.P.TBg[1],
		s.P.Bg[1],
	)

	rl.DrawRectangleGradientV(
		int32(pb.p.X),
		int32(localPriceY+2),
		int32(pb.s.X),
		int32(rh+4),
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
	priceText := fmt.Sprintf("%.5f", cs.Price)
	if len(priceText) > PRICE_BAR_MAX_CONTENT_CAP {
		priceText = priceText[:PRICE_BAR_MAX_CONTENT_CAP]
	}

	localTextY := localPriceY - rh
	rl.DrawTextEx(
		s.F,
		priceText,
		rl.Vector2{X: lable_xPos, Y: localTextY},
		rh,
		0,
		s.P.Fg[1],
	)

	// Cursor price indicator
	if cs.Cursor.Y > 0 {
		localCursorY := cs.Cursor.Y + pb.p.Y - cs.Shift.Y
		localTextY := localCursorY - rowCenter

		priceText := fmt.Sprintf("%.5f", cs.CursorPrice)
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
			rh,
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

// func DrawCenteredGradientV( // TODO
// 	pX int32, centerY int32, width int32, height int32,
// 	topColor, bottomColor rl.Color,
// ) {
// 	halfHeight := height / 2
//
// 	rl.DrawRectangleGradientV(
// 		pX,
// 		centerY-halfHeight,
// 		width,
// 		halfHeight,
// 		topColor,
// 		bottomColor,
// 	)
//
// 	rl.DrawRectangleGradientV(
// 		pX,
// 		centerY,
// 		width,
// 		halfHeight,
// 		bottomColor,
// 		topColor,
// 	)
// }

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

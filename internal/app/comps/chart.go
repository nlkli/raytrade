package comps

import (
	"fmt"
	"math"
	"nlkli/raytrade/internal/app/core"
	"nlkli/raytrade/internal/broker"
	"nlkli/raytrade/internal/cdl"
	"strconv"
	"strings"
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	PRICE_BAR_MAX_NUMBERS_CAP float32 = 7
	PRICE_BAR_MAX_CONTENT_CAP int     = 7 + 1 // Numbers + dot
	PRICE_BAR_LABLE_XPD       float32 = 4     // Padding
	EXTEND_CANDLES_LIMIT              = 100
)

type chartLocalOrder struct {
	price     float64
	priceY    float32
	side      broker.Side
	textInfo  string
	textInfoW float32
}

type chartLocalState struct {
	rh      float32
	rhScale float32

	lastPriceY float32 // Last price y coord

	posEntryPriceY float32

	maxVisPrice float64 // Topmost visible price value
	pxPerPrice  float64 // Conversion factor: pixels per one price unit

	startSec float32 // Unix timestamp (sec) of first visible candle
	secPerPx float32 // Conversion factor: seconds per one pixel

	gridStepV rl.Vector2

	cursorV rl.Vector2

	endCandleIdx int

	localOrder []chartLocalOrder
}

func (s *chartLocalState) priceToWorldY(p float64) float32 {
	return float32((s.maxVisPrice - p) * s.pxPerPrice)
}

func (s *chartLocalState) worldYToPrice(y float32) float64 {
	return s.maxVisPrice - float64(y)/s.pxPerPrice
}

type Chart struct {
	*Rect

	StateIdx int

	localState chartLocalState

	c  Canvas
	tl TimeLine
	pb PriceBar
	cr Crossing
}

func (ch *Chart) R() *Rect {
	return ch.Rect
}

func (ch *Chart) Render(s *core.State) {
	cs := s.Chart[ch.StateIdx] // Chart global state
	ls := &ch.localState       // Chart lacal state

	ls.rh = s.RH - cs.RHD

	if s.WRF || s.RH_Dirty {
		ls.rhScale = ls.rh / float32(s.F.BaseSize)
		var priceBarW float32
		if cs.ShowPriceBar {
			priceBarW = (s.TextNumSV.X*PRICE_BAR_MAX_NUMBERS_CAP+s.TextDotW)*
				ls.rhScale +
				PRICE_BAR_LABLE_XPD*2
		}
		var timeLineH float32
		if cs.ShowTimeLine {
			timeLineH = ls.rh
		}

		l, r := ch.SplitV(ch.s.X - priceBarW)
		ch.c.Rect, ch.tl.Rect = l.SplitH(ch.s.Y - timeLineH)
		ch.pb.Rect, ch.cr.Rect = r.SplitH(ch.s.Y - timeLineH)
	}

	n := len(cs.Candles)
	if n == 0 {
		return
	}

	ch.c.Render(s, cs, ls)
	ch.pb.Render(s, cs, ls)
	ch.tl.Render(s, cs, ls)
	// ch.cr.Render(s, cs)

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
		8,
		s.P.Bg[1],
		s.P.TBg[1],
	)

	if cs.ShowLable && ch.s.X > 200 {
		rl.DrawTextEx(
			s.F,
			cs.LableString,
			rl.Vector2{
				X: ch.p.X + 2,
				Y: ch.p.Y + 2,
			},
			ls.rh,
			0,
			s.P.Comment,
		)
	}

	if ch.c.endCandleIdx >= n && !cs.ExtendCandlesF {
		cs.ExtendCandlesF = true
		s.BTX <- &core.ExtendStartCandles{
			Idx:      ch.StateIdx,
			Category: cs.Category,
			Symbol:   cs.Symbol,
			Interval: cs.Interval,
			Candles:  cs.Candles,
			Limit:    EXTEND_CANDLES_LIMIT,
		}
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

	endCandleIdx int
}

func (c *Canvas) Render(s *core.State, cs *core.ChartState, ls *chartLocalState) {
	candles := cs.Candles
	n := len(candles)

	sf := cs.Scale
	cw := cs.CW * sf.X  // candle width * scale
	stepX := cw + cs.CG // total width including gap

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
		ls.maxVisPrice = cs.MidP + visPriceRng*0.5
		ls.pxPerPrice = float64(c.s.Y) / visPriceRng

		priceStep := quantizePriceStep(
			float64(s.RH * 2.5 * float32(1/ls.pxPerPrice)),
		)
		ls.gridStepV.Y = float32(priceStep * ls.pxPerPrice)

		// X-axis grid calculation
		timeSecRng := cs.Cap * cs.SecInterval
		ls.startSec = float32(visC[len(visC)-1].Time)/1000 - timeSecRng
		ls.secPerPx = timeSecRng / c.s.X

		timeStep := quantizeTimeStep(s.RH * 6 * ls.secPerPx)
		ls.gridStepV.X = timeStep * (1 / ls.secPerPx)

		ps := strconv.FormatFloat(candles[n-1].O, 'f', -1, 64)
		pp := len(ps) - (strings.IndexByte(ps, '.') + 1)
		cs.PricePrecision = max(cs.PricePrecision, pp)
		ps = strconv.FormatFloat(candles[n-1].C, 'f', -1, 64)
		pp = len(ps) - (strings.IndexByte(ps, '.') + 1)
		cs.PricePrecision = max(cs.PricePrecision, pp)
	} else {
		visPriceRng := cs.RngP / float64(sf.Y)
		ls.maxVisPrice = cs.MidP + visPriceRng*0.5
		ls.pxPerPrice = float64(c.s.Y) / visPriceRng
	}

	// Current price position
	ls.lastPriceY = ls.priceToWorldY(cs.LastPrice)

	// Draw grid
	if cs.ShowGrid {
		// Horizontal grid lines
		for y := c.s.Y - ls.gridStepV.Y*.5; y > 0; y -= ls.gridStepV.Y {
			localY := y + c.p.Y

			rl.DrawLineEx(
				rl.Vector2{X: c.p.X, Y: localY},
				rl.Vector2{X: c.p.X + c.s.X, Y: localY},
				1,
				s.P.Bg[2],
			)
		}
		// Vertical grid lines
		for x := c.s.X - ls.gridStepV.X*.5; x > 0; x -= ls.gridStepV.X {
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
			levelY := ls.priceToWorldY(l)

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

			// if worldMouse.X < -10 ||
			// 	worldMouse.Y < -10 ||
			// 	worldMouse.X > c.s.X+10 ||
			// 	worldMouse.Y > c.s.Y+10 {
			// 	cs.Lines = cs.Lines[:len(cs.Lines)-1]
			// 	cs.IsLineDuring = false
			// }

			l := lines[len(lines)-1]

			lineX0 := (l[0].X - ls.startSec) / ls.secPerPx
			lineY0 := ls.priceToWorldY(float64(l[0].Y))

			rl.DrawLineEx(
				rl.Vector2{X: lineX0, Y: lineY0},
				worldMouse,
				1,
				s.P.Dim.Blue,
			)

			diff := cs.CursorPrice - float64(l[0].Y)

			var pDiff float64
			if diff != 0 {
				pDiff = diff / float64(cs.CursorPrice) * 100
			}

			pDiffText := strconv.FormatFloat(pDiff, 'f', 2, 64)

			var color rl.Color
			if pDiff > 0 {
				color = s.P.Base.Cyan
			} else {
				color = s.P.Base.Magenta
			}

			rl.DrawTextEx(
				s.F,
				pDiffText,
				rl.Vector2{
					X: worldMouse.X + 2,
					Y: worldMouse.Y - ls.rh,
				},
				ls.rh,
				0,
				color,
			)

			lines = lines[:len(lines)-1]
		}

		for _, l := range lines {
			lineX0 := (l[0].X - ls.startSec) / ls.secPerPx
			lineX1 := (l[1].X - ls.startSec) / ls.secPerPx

			lineY0 := ls.priceToWorldY(float64(l[0].Y))
			lineY1 := ls.priceToWorldY(float64(l[1].Y))

			rl.DrawLineEx(
				rl.Vector2{X: lineX0, Y: lineY0},
				rl.Vector2{X: lineX1, Y: lineY1},
				1,
				s.P.Dim.Blue,
			)
		}
	}

	// Position
	if cs.ShowPosition &&
		cs.PositionIdx >= 0 && cs.PositionIdx < len(s.Position.List) {

		pos := s.Position.List[cs.PositionIdx]

		ls.posEntryPriceY = ls.priceToWorldY(pos.EntryPrice)

		textInfo := fmt.Sprintf(
			"P|%.2f",
			pos.UnrealisedPnl,
		)

        textInfoW := rl.MeasureTextEx(s.F, textInfo, ls.rh, 0).X

		var color rl.Color
		var colorT rl.Color
		if pos.Side == broker.Long {
			color = s.P.Diff.Add
			colorT = s.P.Bright.Green
		} else {
			color = s.P.Diff.Delete
			colorT = s.P.Bright.Red
		}

		rl.DrawTextEx(
			s.F,
			textInfo,
			rl.Vector2{X: cs.Shift.X + 4, Y: ls.posEntryPriceY - ls.rh*0.5},
			ls.rh,
			0,
			colorT,
		)

		rl.DrawLineEx(
			rl.Vector2{X: cs.Shift.X + textInfoW + 6, Y: ls.posEntryPriceY},
			rl.Vector2{X: c.s.X + cs.Shift.X, Y: ls.posEntryPriceY},
			1,
			color,
		)
	}

	if cs.ShowOrder {

		if s.Order.Forced {
			ls.localOrder = ls.localOrder[:0]
			for i, oi := range s.Order.List {
				if oi.Category != cs.Category ||
					oi.Symbol != cs.Symbol ||
					oi.Price <= 0 {
					continue
				}

				orderPriceY := ls.priceToWorldY(oi.Price)

				textInfo := fmt.Sprintf(
					"%d|%s",
					i, strconv.FormatFloat(oi.LeavesQty, 'f', -1, 64),
				)

				textInfoW := rl.MeasureTextEx(s.F, textInfo, ls.rh, 0).X

				ls.localOrder = append(
					ls.localOrder,
					chartLocalOrder{
						price:     oi.Price,
						priceY:    orderPriceY,
						side:      oi.Side,
						textInfo:  textInfo,
						textInfoW: textInfoW,
					},
				)
			}
		}

		for _, oi := range ls.localOrder {
			var color rl.Color
			var colorT rl.Color
			if oi.side == broker.Long {
				color = s.P.Diff.Add
				colorT = s.P.Bright.Green
			} else {
				color = s.P.Diff.Delete
				colorT = s.P.Bright.Red
			}

			rl.DrawTextEx(
				s.F,
				oi.textInfo,
				rl.Vector2{X: cs.Shift.X + 4, Y: oi.priceY - ls.rh*0.5},
				ls.rh,
				0,
				colorT,
			)

			rl.DrawLineEx(
				rl.Vector2{X: cs.Shift.X + oi.textInfoW + 6, Y: oi.priceY},
				rl.Vector2{X: c.s.X + cs.Shift.X, Y: oi.priceY},
				1,
				color,
			)

		}

	}

	// Current price line
	rl.DrawLineEx(
		rl.Vector2{X: cs.Shift.X, Y: ls.lastPriceY},
		rl.Vector2{X: c.s.X + cs.Shift.X, Y: ls.lastPriceY},
		1,
		s.P.Bg[4],
	)

	start := 1
	if cs.Shift.X < 0 {
		start -= int(cs.Shift.X / stepX)
	}
	c.endCandleIdx = min(n, int(cs.Cap)+start+2)

	// Draw candles from right to left (newest to oldest)
	for i := start; i < c.endCandleIdx; i++ {
		candle := candles[n-i]

		xPos := c.s.X - stepX*float32(i)

		yO := ls.priceToWorldY(candle.O)
		yH := ls.priceToWorldY(candle.H)
		yL := ls.priceToWorldY(candle.L)
		yC := ls.priceToWorldY(candle.C)

		color := s.P.Base.Green
		if candle.C < candle.O {
			color = s.P.Base.Red
		}

		// Draw wick
		wickX := xPos + (cw * 0.5)
		rl.DrawLineEx(
			rl.Vector2{X: wickX, Y: yH},
			rl.Vector2{X: wickX, Y: yL},
			cs.CWW,
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
		cs.CursorPrice = ls.worldYToPrice(cs.Cursor.Y)

		if s.E.Mouse.Click[0] {
			ratio := math.Pow(10, float64(cs.PricePrecision))
			rp := math.Round(cs.CursorPrice*ratio) / ratio
			s.BTX <- &core.SelectInstrumentPrice{
				Category: cs.Category,
				Symbol:   cs.Symbol,
				Price:    rp,
			}
		}

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

		if !cs.ShowPriceBar {
			priceText := fmt.Sprintf("%.5f", cs.CursorPrice)
			if len(priceText) > PRICE_BAR_MAX_CONTENT_CAP {
				priceText = priceText[:PRICE_BAR_MAX_CONTENT_CAP]
			}
			priceTextW := s.TextNumSV.X*
				float32(len(priceText))*ls.rhScale
			rl.DrawTextEx(
				s.F,
				priceText,
				rl.Vector2{
					X: c.s.X + cs.Shift.X - priceTextW,
					Y: wm.Y - ls.rh,
				},
				ls.rh,
				0,
				s.P.Base.Orange,
			)

		}

		if s.E.Mouse.HoldLeft {

			cs.Shift.X -= s.E.Mouse.Delta.X
			cs.Shift.Y -= s.E.Mouse.Delta.Y

		} else {

			if s.E.Mouse.Click[0] && cs.IsLineDuring && len(cs.Lines) > 0 {

				cs.Lines[len(cs.Lines)-1][1] = rl.Vector2{
					X: ls.startSec + wm.X*ls.secPerPx,
					Y: float32(
						ls.worldYToPrice(wm.Y),
					),
				}

				cs.IsLineDuring = false
			}

			if s.E.Mouse.DoubleClick && !cs.IsLineDuring {
				cs.Lines = append(
					cs.Lines,
					[2]rl.Vector2{
						{
							X: ls.startSec + wm.X*ls.secPerPx,
							Y: float32(
								ls.worldYToPrice(wm.Y),
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

func (tl *TimeLine) Render(s *core.State, cs *core.ChartState, ls *chartLocalState) {
	if tl.s.Y == 0 {
		return
	}
	// Text offset for centering
	halfText := (s.TextNumSV.X*2 + s.TextDotW/2) * ls.rhScale

	rl.BeginScissorMode(
		int32(tl.p.X),
		int32(tl.p.Y),
		int32(tl.s.X),
		int32(tl.s.Y),
	)

	// Draw time labels at grid positions
	for x := tl.s.X - ls.gridStepV.X*.5; x > 0; x -= ls.gridStepV.X {
		localX := x + tl.p.X - halfText

		xSec := (x + cs.Shift.X) * ls.secPerPx
		tUnix := time.Unix(int64(ls.startSec+xSec), 0)
		timeText := tUnix.Format("15:04")

		rl.DrawTextEx(
			s.F,
			timeText,
			rl.Vector2{X: localX, Y: tl.p.Y},
			ls.rh,
			0,
			s.P.Fg[3],
		)
	}

	rl.EndScissorMode()
}

type PriceBar struct {
	*Rect
}

func (pb *PriceBar) Render(s *core.State, cs *core.ChartState, ls *chartLocalState) {
	if pb.s.X == 0 {
		return
	}

	halfRH := ls.rh * .5
	// Lables offset x (padding)
	lable_xPos := pb.p.X + PRICE_BAR_LABLE_XPD

	rl.BeginScissorMode(
		int32(pb.p.X),
		int32(pb.p.Y),
		int32(pb.s.X),
		int32(pb.s.Y),
	)

	// Draw price labels at grid positions
	for y := pb.s.Y - ls.gridStepV.Y*.5; y > 0; y -= ls.gridStepV.Y {
		localTextY := y + pb.p.Y - halfRH

		yPrice := ls.worldYToPrice(y + cs.Shift.Y)

		priceText := fmt.Sprintf("%.5f", yPrice)
		if len(priceText) > PRICE_BAR_MAX_CONTENT_CAP {
			priceText = priceText[:PRICE_BAR_MAX_CONTENT_CAP]
		}

		rl.DrawTextEx(
			s.F,
			priceText,
			rl.Vector2{X: lable_xPos, Y: localTextY},
			ls.rh,
			0,
			s.P.Fg[3],
		)
	}

	// Position entry price lable
	if cs.PositionIdx >= 0 && len(s.Position.List) > cs.PositionIdx {
		pos := s.Position.List[cs.PositionIdx]

		priceText := fmt.Sprintf("%.5f", pos.EntryPrice)
		if len(priceText) > PRICE_BAR_MAX_CONTENT_CAP {
			priceText = priceText[:PRICE_BAR_MAX_CONTENT_CAP]
		}

		localPosEP := ls.posEntryPriceY + pb.p.Y - cs.Shift.Y
		localTextY := localPosEP - halfRH

		// Gradient background for position entry price
		DrawGradientAround(
			int32(localPosEP),
			int32(pb.p.X),
			int32(halfRH)+3,
			int32(pb.s.X),
			s.P.Bg[1],
			s.P.TBg[1],
		)

		rl.DrawTextEx(
			s.F,
			priceText,
			rl.Vector2{X: lable_xPos, Y: localTextY},
			ls.rh,
			0,
			s.P.Fg[1],
		)
	}

	// Current price position in local coordinates
	localPriceY := ls.lastPriceY + pb.p.Y - cs.Shift.Y

	// Gradient around current price line
	DrawGradientAround(
		int32(localPriceY),
		int32(pb.p.X),
		int32(ls.rh)+3,
		int32(pb.s.X),
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
	priceText := fmt.Sprintf("%.5f", cs.LastPrice)
	if len(priceText) > PRICE_BAR_MAX_CONTENT_CAP {
		priceText = priceText[:PRICE_BAR_MAX_CONTENT_CAP]
	}

	localTextY := localPriceY - ls.rh
	rl.DrawTextEx(
		s.F,
		priceText,
		rl.Vector2{X: lable_xPos, Y: localTextY},
		ls.rh,
		0,
		s.P.Fg[1],
	)

	// Cursor price indicator
	if cs.Cursor.Y != 0 {
		localCursorY := cs.Cursor.Y + pb.p.Y - cs.Shift.Y
		localTextY := localCursorY - halfRH

		priceText := fmt.Sprintf("%.5f", cs.CursorPrice)
		if len(priceText) > PRICE_BAR_MAX_CONTENT_CAP {
			priceText = priceText[:PRICE_BAR_MAX_CONTENT_CAP]
		}

		DrawGradientAround(
			int32(localCursorY),
			int32(pb.p.X),
			int32(halfRH)+3,
			int32(pb.s.X),
			s.P.Bg[1],
			s.P.TBg[1],
		)

		rl.DrawTextEx(
			s.F,
			priceText,
			rl.Vector2{X: lable_xPos, Y: localTextY},
			ls.rh,
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

func DrawGradientAround(centerY int32, px, sy, sx int32, color1, color2 rl.Color) {
	rl.DrawRectangleGradientV(
		px,
		centerY-sy,
		sx,
		sy,
		color2,
		color1,
	)

	rl.DrawRectangleGradientV(
		px,
		centerY,
		sx,
		sy,
		color1,
		color2,
	)
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

package comps

import (
	"fmt"
	"math"
	"nlkli/raytrade/internal/app/core"
	"nlkli/raytrade/internal/broker"
	"strconv"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	POSITION_CONTENT_PDX float32 = 4 // padding
)

type Position struct {
	*Rect

	sideTextW float32
}

func (p *Position) R() *Rect {
	return p.Rect
}

func (p *Position) Render(s *core.State) {
	ps := s.Position

	rh1 := s.RH - ps.RHD

	if s.WRF || s.RH_Dirty {
		p.sideTextW = s.TextNumSV.X * 5 * rh1 / float32(s.F.BaseSize)
	}

	if len(ps.List) == 0 {
		return
	}

	rl.BeginScissorMode(
		int32(p.p.X),
		int32(p.p.Y),
		int32(p.s.X),
		int32(p.s.Y),
	)

	var offsetY float32
	cursor := rl.Vector2{
		X: p.p.X + POSITION_CONTENT_PDX, Y: p.p.Y + offsetY,
	}
	for _, pi := range ps.List {

		rl.DrawRectangleV(
			rl.Vector2{X: p.p.X, Y: cursor.Y},
			rl.Vector2{X: p.s.X, Y: s.RH},
			s.P.Bg[0],
		)

		rl.DrawTextEx(
			s.F,
			pi.Symbol,
			cursor,
			s.RH,
			0,
			s.P.Dim.Yellow,
		)

		cursor.Y += s.RH

		if pi.EntryPrice == 0 {
			continue
		}

		var sideText string
		if pi.Side == broker.Long {
			sideText = "Long"
		} else {
			sideText = "Short"
		}

		var sColor rl.Color
		if pi.Side == broker.Long {
			sColor = s.P.Dim.Green
		} else {
			sColor = s.P.Dim.Red
		}

		rl.DrawTextEx(
			s.F,
			sideText,
			cursor,
			rh1,
			0,
			sColor,
		)

		rl.DrawTextEx(
			s.F,
			fmt.Sprintf("X%d", pi.Leverage),
			rl.Vector2{
				X: cursor.X + p.sideTextW,
				Y: cursor.Y,
			},
			rh1,
			0,
			s.P.Fg[1],
		)

		cursor.Y += rh1

		var roi float64
		if pi.PositionIM != 0 {
			roi = (pi.UnrealisedPnl / pi.PositionIM) * 100
			roi = math.Round(roi*100) / 100
		}
		profitText := fmt.Sprintf("%.2f (%.2f%%)", pi.UnrealisedPnl, roi)

		var pColor rl.Color
		if roi > 0 {
			pColor = s.P.Base.Green
		} else {
			pColor = s.P.Base.Red
		}

		rl.DrawTextEx(
			s.F,
			profitText,
			cursor,
			s.RH,
			0,
			pColor,
		)

		cursor.Y += s.RH

		qtyInfo := fmt.Sprintf(
			"S: %s (%.2f)",
			strconv.FormatFloat(pi.Size, 'f', -1, 64),
			pi.PositionValue,
		)

		rl.DrawTextEx(
			s.F,
			qtyInfo,
			cursor,
			rh1,
			0,
			s.P.Fg[1],
		)

		cursor.Y += rh1

		var priceDiff float64
		if pi.EntryPrice != 0 {
			priceDiff = (pi.MarkPrice - pi.EntryPrice) / pi.EntryPrice * 100
			priceDiff = math.Round(priceDiff*100) / 100
		}
		priceInfo := fmt.Sprintf(
			"P: %.6f (%.2f%%)",
			pi.EntryPrice,
			priceDiff,
		)

		rl.DrawTextEx(
			s.F,
			priceInfo,
			cursor,
			rh1,
			0,
			s.P.Fg[1],
		)

		cursor.Y += rh1

		if pi.TakeProfit != 0 {
			takeProfitText := fmt.Sprintf(
				"TP: %s",
				strconv.FormatFloat(pi.TakeProfit, 'f', -1, 64),
			)

			rl.DrawTextEx(
				s.F,
				takeProfitText,
				cursor,
				rh1,
				0,
				s.P.Dim.Green,
			)

			cursor.Y += rh1
		}

		if pi.StopLoss != 0 {
			stopLossText := fmt.Sprintf(
				"SL: %s",
				strconv.FormatFloat(pi.StopLoss, 'f', -1, 64),
			)

			rl.DrawTextEx(
				s.F,
				stopLossText,
				cursor,
				rh1,
				0,
				s.P.Dim.Red,
			)

			cursor.Y += rh1
		}
	}

	rl.DrawLineEx(
		rl.Vector2{X: p.p.X, Y: cursor.Y},
		rl.Vector2{X: p.p.X + p.s.X, Y: cursor.Y},
		1,
		s.P.Bg[0],
	)

	rl.EndScissorMode()

	p.Outline(1, s.P.Bg[0])
}

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

type positionForcedData struct {
	headerInfo string
	roi        float64
	profitInfo string
	qtyInfo    string
	priceInfo  string
	tpInfo     string
	slInfo     string
	height     float32
}

type Position struct {
	*Rect

	forcedData []positionForcedData
}

func (p *Position) R() *Rect {
	return p.Rect
}

func (p *Position) Render(s *core.State) {
	ps := &s.Position

	rh1 := s.RH - ps.RHD
	rhScale := rh1 / float32(s.F.BaseSize)

	if len(ps.List) == 0 {
		return
	}

	if ps.Forced {
		p.forcedData = p.forcedData[:0]

		for i, pi := range ps.List {
			var fd positionForcedData

			fd.headerInfo = fmt.Sprintf(
				"%d.%s.%s", i, pi.Category.AsString(true), pi.Symbol,
			)

			if pi.PositionIM != 0 {
				fd.roi = (pi.UnrealisedPnl / pi.PositionIM) * 100
				fd.roi = math.Round(fd.roi*100) / 100
			}
			fd.profitInfo = fmt.Sprintf("%.2f (%.2f%%)", pi.UnrealisedPnl, fd.roi)

			fd.qtyInfo = fmt.Sprintf(
				"S: %s (%.2f)",
				strconv.FormatFloat(pi.Size, 'f', -1, 64),
				pi.PositionValue,
			)

			var priceDiff float64
			if pi.EntryPrice != 0 {
				priceDiff = (pi.MarkPrice - pi.EntryPrice) / pi.EntryPrice * 100
				priceDiff = math.Round(priceDiff*100) / 100
			}
			fd.priceInfo = fmt.Sprintf(
				"P: %.6f (%.2f%%)",
				pi.EntryPrice,
				priceDiff,
			)

			fd.height = s.RH*2 + rh1*3

			if pi.TakeProfit != 0 {
				fd.tpInfo = fmt.Sprintf(
					"TP: %s",
					strconv.FormatFloat(pi.TakeProfit, 'f', -1, 64),
				)
				fd.height += rh1
			}

			if pi.StopLoss != 0 {
				fd.slInfo = fmt.Sprintf(
					"SL: %s",
					strconv.FormatFloat(pi.StopLoss, 'f', -1, 64),
				)
				fd.height += rh1
			}
			p.forcedData = append(
				p.forcedData,
				fd,
			)
		}

		ps.Forced = false
	}

	rl.BeginScissorMode(
		int32(p.p.X),
		int32(p.p.Y),
		int32(p.s.X),
		int32(p.s.Y),
	)

	cursor := rl.Vector2{
		X: p.p.X + POSITION_CONTENT_PDX, Y: p.p.Y + ps.OffsetY,
	}
	for i, pi := range ps.List {

		if pi.EntryPrice == 0 {
			continue
		}

		fd := p.forcedData[i]

		rl.DrawRectangleV(
			rl.Vector2{X: p.p.X, Y: cursor.Y},
			rl.Vector2{X: p.s.X, Y: s.RH},
			s.P.Bg[0],
		)

		rl.DrawTextEx(
			s.F,
			fd.headerInfo,
			cursor,
			s.RH,
			0,
			s.P.Dim.Yellow,
		)

		cursor.Y += s.RH

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
				X: cursor.X + s.TextNumSV.X*
					float32(len(sideText)+1)*rhScale,
				Y: cursor.Y,
			},
			rh1,
			0,
			s.P.Fg[1],
		)

		cursor.Y += rh1

		var pColor rl.Color
		if fd.roi > 0 {
			pColor = s.P.Base.Green
		} else {
			pColor = s.P.Base.Red
		}

		rl.DrawTextEx(
			s.F,
			fd.profitInfo,
			cursor,
			s.RH,
			0,
			pColor,
		)

		cursor.Y += s.RH

		rl.DrawTextEx(
			s.F,
			fd.qtyInfo,
			cursor,
			rh1,
			0,
			s.P.Fg[1],
		)

		cursor.Y += rh1

		rl.DrawTextEx(
			s.F,
			fd.priceInfo,
			cursor,
			rh1,
			0,
			s.P.Fg[1],
		)

		cursor.Y += rh1

		if pi.TakeProfit != 0 {
			rl.DrawTextEx(
				s.F,
				fd.tpInfo,
				cursor,
				rh1,
				0,
				s.P.Dim.Green,
			)

			cursor.Y += rh1
		}

		if pi.StopLoss != 0 {
			rl.DrawTextEx(
				s.F,
				fd.slInfo,
				cursor,
				rh1,
				0,
				s.P.Dim.Red,
			)

			cursor.Y += rh1
		}
	}

	if !s.E.Mouse.Captured && p.Rect.ContainsV(s.E.Mouse.Pos) {
		if s.E.Mouse.HoldLeft {
			ps.OffsetY -= s.E.Mouse.Delta.Y
		}

		if s.E.Mouse.Click[0] && s.E.Mouse.Pos.Y >= p.p.Y+ps.OffsetY {
			cursorY := p.p.Y + ps.OffsetY
			for i := range ps.List {
				fd := p.forcedData[i]
				cursorY += fd.height
				if s.E.Mouse.Pos.Y < cursorY {
					s.BTX <- &core.SelectPosition{
						Position: &ps.List[i],
					}
					break
				}
			}
		}

		s.E.Mouse.Captured = true
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

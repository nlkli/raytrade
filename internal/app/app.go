package app

import (
	"fmt"
	"math"
	"nlkli/raytrade/internal/cdl"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type Chart struct {
	scale, shift rl.Vector2
	maxV, minV float32
}

func NewChart() *Chart {
	return &Chart{}
}

func (c *Chart) Scale(factor float32) {
	c.scale = rl.Vector2AddValue(c.scale, factor)
}

func (c *Chart) DrawRectangleV(position rl.Vector2, size rl.Vector2, color rl.Color) {
	rl.DrawRectangleV(rl.Vector2Add(position, c.shift), rl.Vector2Add(size, c.scale), color)
}

func PriceToY(price, maxV, minV, rY, rHeight float32) float32 {
	return rY + (maxV - price) / (maxV - minV) * rHeight
}

func (c *Chart) DrawCandles(candles []cdl.Candle) {
	if len(candles) == 0 {
		return
	}

	const width float32 = 5.
	const gap float32 = 1.

	maxV, minV := cdl.MinMaxPrice(candles)
	c.maxV, c.minV = float32(maxV), float32(minV)

	winSize := rl.NewVector2(float32(rl.GetScreenWidth()), float32(rl.GetScreenHeight()))

	n := len(candles) - 1
	for i := range candles {
		candle := candles[n-i]

		x := winSize.X - (width + gap) * float32(i + 1)
		if x < 0 {
			return
		}

		yO := PriceToY(float32(candle.O), c.maxV, c.minV, 0, winSize.Y)
        yC := PriceToY(float32(candle.C), c.maxV, c.minV, 0, winSize.Y)
        yH := PriceToY(float32(candle.H), c.maxV, c.minV, 0, winSize.Y)
        yL := PriceToY(float32(candle.L), c.maxV, c.minV, 0, winSize.Y)

		var color rl.Color 
		if candle.C >= candle.O {
			color = rl.Green
		} else {
			color = rl.Red
		}

		halfW := x + width / 2
		rl.DrawLineV(rl.NewVector2(halfW, yH), rl.NewVector2(halfW, yL), color)
        rl.DrawRectangleV(
			rl.NewVector2(x, min(yO, yC)),
			rl.NewVector2(width, float32(max(1, math.Abs(float64(yO - yC))))),
            color,
        )

	}
}

func Run() error {
	candles, err := cdl.CandlesFromCsv("BTCUSDT-15m-2025-12.csv")
	if err != nil {
		return err
	}

	fmt.Println(candles[0])

	rl.SetConfigFlags(rl.FlagWindowResizable)
	rl.InitWindow(800, 600, "raytrade")
	defer rl.CloseWindow()

	rl.SetTargetFPS(60)

	chart := NewChart()

	for !rl.WindowShouldClose() {
		if rl.IsKeyDown(rl.KeyLeft) {
			chart.shift.X -= .9
		}
		if rl.IsKeyDown(rl.KeyRight) {
			chart.shift.X += .9
		}
		if rl.IsKeyDown(rl.KeyUp) {
			chart.shift.Y -= .9
		}
		if rl.IsKeyDown(rl.KeyDown) {
			chart.shift.Y += .9
		}
		if rl.IsKeyDown(rl.KeyEqual) {
			chart.Scale(.3)
		}
		if rl.IsKeyDown(rl.KeyMinus) {
			chart.Scale(-.3)
		}

		rl.BeginDrawing()

		rl.ClearBackground(rl.Black)

		// mPos := rl.GetMousePosition()

        r := rl.IsWindowResized()
		rl.DrawText(fmt.Sprintf("%d %d %+v", rl.GetScreenWidth(), rl.GetScreenHeight(), r), 2, 2, 9, rl.White)

		// chart.DrawRectangleV(rl.NewVector2(10, 10), rl.NewVector2(20, 20), rl.White)
		// rl.DrawRectangleV(mPos, rl.NewVector2(10, 10), rl.NewColor(255, 255, 0, 255))
		chart.DrawCandles(candles[len(candles)-666:])

		rl.EndDrawing()
	}

	return nil
}

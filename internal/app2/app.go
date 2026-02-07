package app2

import (
	"context"
	"encoding/json"
	"nlkli/raytrade/internal/broker"
	"nlkli/raytrade/internal/broker/bybit"
	"nlkli/raytrade/internal/cdl"
	"os"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type App struct {
	state      *State
	root       *Root
	worker     *Worker
	controller *Controller

	ctx context.Context
}

func (a *App) Frame() {

	a.state.WHF = rl.IsWindowHidden()
	a.state.WFF = rl.IsWindowFocused()
	a.state.WRF = rl.IsWindowResized() || a.state.FN == 0

	if a.state.WRF {
		sW, sH := rl.GetScreenWidth(), rl.GetScreenHeight()
		a.state.WS = rl.NewVector2(float32(sW), float32(sH))
	}

	a.state.E = a.controller.Event(a.state.M)

	rl.BeginDrawing()
	a.root.Render(a.state)
	rl.EndDrawing()

	a.state.FN += 1

	select {
	case f := <-a.worker.Rx:
		if f != nil {
			f(a.state)
		}
	default:
		return
	}
}

func Run(ctx context.Context, configPath string) error {
	b, err := os.ReadFile(configPath)
	if err != nil {
		return err
	}

	var c Config
	if err = json.Unmarshal(b, &c); err != nil {
		return err
	}

	client := bybit.NewClientFromEnv(ctx)
	br := bybit.NewBroker(client)

	state := InitState(&c)
	worker := NewWorker(ctx, br)
	state.WTX = worker.Tx

	go func() {
		candles, err := br.GetCandles(broker.Futures, "FARTCOINUSDT", cdl.M1, 100, nil, nil)
		if err != nil {
			return
		}
		state.Chart.candles = candles
	}()

	app := &App{
		state:      state,
		root:       InitRoot(),
		worker:     worker,
		controller: NewController(),
	}

	rl.SetConfigFlags(rl.FlagWindowResizable)
	rl.InitWindow(c.InitWindow.Width, c.InitWindow.Height, c.InitWindow.Title)
	rl.SetExitKey(0)

	defer rl.CloseWindow()

	rl.SetTargetFPS(c.TargetFPS)

	for !rl.WindowShouldClose() {
		app.Frame()
	}

	return nil
}

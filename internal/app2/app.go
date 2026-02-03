package app2

import (
	"context"
	"encoding/json"
	"os"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type App struct {
	state  *State
	root   *Root
	worker *Worker

	ctx context.Context
}

func (a *App) Frame() {
	a.state.WHF = rl.IsWindowHidden()
	a.state.WFF = rl.IsWindowFocused()
	a.state.WRF = rl.IsWindowResized()

	if a.state.WRF {
		sW, sH := rl.GetScreenWidth(), rl.GetScreenHeight()
		a.state.W = rl.NewVector2(float32(sW), float32(sH))
	}

	cp := rl.GetCharPressed()
	a.state.CPF = cp != 0
	a.state.CP = rune(cp)

	rl.BeginDrawing()
	a.root.Render(a.state)
	rl.EndDrawing()

	a.state.FN += 1

	select {
	case f := <-a.worker.Rx:
		f(a.state)
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

	state := InitState(&c)
	worker := NewWorker(ctx)
	state.WTX = worker.Tx

	app := &App{
		state:  state,
		root:   InitRoot(),
		worker: worker,
	}

	rl.SetConfigFlags(rl.FlagWindowResizable)
	rl.InitWindow(c.InitWindow.Width, c.InitWindow.Height, c.InitWindow.Title)
	rl.SetExitKey(0)

	defer rl.CloseWindow()

	rl.SetTargetFPS(60)

	for !rl.WindowShouldClose() {
		app.Frame()
	}

	return nil
}

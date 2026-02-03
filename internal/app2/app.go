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

	ctx    context.Context
	cancel context.CancelFunc
}

func (a *App) Frame() {
	rl.BeginDrawing()
	rl.ClearBackground(a.state.P.Bg[1])

	a.state.WR = rl.IsWindowResized()
	if a.state.WR {
		sW, sH := rl.GetScreenWidth(), rl.GetScreenHeight()
		a.state.W = rl.NewVector2(float32(sW), float32(sH))
	}
	cp := rl.GetCharPressed()
	if cp != 0 {
		a.state.PF = true
	} else {
		a.state.PF = false
	}
	a.state.CP = rune(cp)

	a.root.Render(a.state)

	a.state.FN += 1

	rl.EndDrawing()
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

	app := &App{
		state: InitState(&c),
		root:  InitRoot(),
	}

	rl.SetConfigFlags(rl.FlagWindowResizable)
	rl.InitWindow(c.InitWindow.Width, c.InitWindow.Height, c.InitWindow.Title)

	defer rl.CloseWindow()

	rl.SetTargetFPS(60)

	for !rl.WindowShouldClose() {
		app.Frame()
	}

	return nil
}

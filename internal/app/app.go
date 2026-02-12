package app

import (
	"context"
	"encoding/json"
	"nlkli/raytrade/internal/broker/bybit"
	"os"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type App struct {
	S    *State
	Root *Root
	CMD  *CMD
	BG   *Background
	C    *Controller

	ctx context.Context
}

func (a *App) Frame() {
	// a.state.FT = time.Duration(
	// 	rl.GetFrameTime() * float32(time.Second),
	// )

	a.S.WHF = rl.IsWindowHidden()
	a.S.WFF = rl.IsWindowFocused()
	a.S.WRF = rl.IsWindowResized() || a.S.FN == 0
	if a.S.WRF {
		sW, sH := rl.GetScreenWidth(), rl.GetScreenHeight()
		a.S.WS = rl.NewVector2(float32(sW), float32(sH))
	}

	a.C.Event(a.S)
	a.BG.Update(a.S)

	rl.BeginDrawing()
	a.Root.Render(a.S)
	rl.EndDrawing()

	a.S.FN += 1

	a.CMD.Update(a.S)
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

	client := bybit.NewClientFromEnv(context.Background())
	br := bybit.NewBroker(client)

	rl.SetConfigFlags(rl.FlagWindowResizable)
	rl.InitWindow(c.InitWindow.Width, c.InitWindow.Height, c.InitWindow.Title)
	rl.SetExitKey(0)

	state := InitState(&c)

	cmd := InitCMD(context.Background())
	state.CMDTX = cmd.Tx
	bg := InitBackground(context.Background(), br)
	state.BTX = bg.Tx

	app := &App{
		S:    state,
		Root: InitRoot(),
		CMD:  cmd,
		BG:   bg,
		C:    NewController(),
	}

	defer rl.CloseWindow()

	rl.SetTargetFPS(c.TargetFPS)

	for !rl.WindowShouldClose() {
		app.Frame()
	}

	return nil
}

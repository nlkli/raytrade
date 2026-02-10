package app

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
	S          *State
	Root       *Root
	BG         *Background
	C *Controller

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

	rl.BeginDrawing()
	a.Root.Render(a.S)
	rl.EndDrawing()

	a.S.FN += 1

	select {
	case f := <-a.BG.Rx:
		if f != nil {
			f(a.S)
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

	rl.SetConfigFlags(rl.FlagWindowResizable)
	rl.InitWindow(c.InitWindow.Width, c.InitWindow.Height, c.InitWindow.Title)

	rl.SetExitKey(0)

	state := InitState(&c)
	// state.CommandLine.Lines = []string{"HELLO", "WORLD", "1", "00000000000000000000", "HELLO", "WORLD", "1", "00000000000000000000", "7777", "--=-=-=-=---=-==--=----=="}
	// TEMP
	candles, err := br.GetCandles(broker.Futures, "BTCUSDT", cdl.M1, 200, nil, nil)
	if err != nil {
		return err
	}
	state.Chart.Candles = candles

	bg := NewBackground(ctx, br)
	state.BTX = bg.Tx

	app := &App{
		S:          state,
		Root:       InitRoot(),
		BG:         bg,
		C: NewController(),
	}

	defer rl.CloseWindow()

	rl.SetTargetFPS(c.TargetFPS)

	for !rl.WindowShouldClose() {
		app.Frame()
	}

	return nil
}

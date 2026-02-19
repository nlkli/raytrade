package app

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"nlkli/raytrade/internal/app/comps"
	"nlkli/raytrade/internal/app/core"
	"nlkli/raytrade/internal/broker/bybit"
	"os"
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type App struct {
	Root *comps.Root
	S    *core.State
	CMD  *core.CMD
	BG   *core.Background
	C    *core.Controller
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

	var c core.Config
	if err = json.Unmarshal(b, &c); err != nil {
		return err
	}

	httpClient := &http.Client{
		Transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout:   3 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			TLSHandshakeTimeout:   3 * time.Second,
			ResponseHeaderTimeout: 5 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
			IdleConnTimeout:       90 * time.Second,
			MaxIdleConns:          100,
			MaxIdleConnsPerHost:   20,
		},
	}

	client := bybit.NewClientFromEnv(
		context.Background(), bybit.WithHttpClient(httpClient),
	)
	br := bybit.NewBroker(client)

	rl.SetConfigFlags(rl.FlagWindowResizable)

	rl.InitWindow(c.InitWindow.Width, c.InitWindow.Height, c.InitWindow.Title)
	rl.SetTraceLogLevel(rl.LogAll)

	rl.SetExitKey(0)

	state := core.InitState(&c)

	cmd := core.InitCMD(context.Background())

	state.CMDTX = cmd.Tx

	bg := core.InitBackground(context.Background(), br)
	state.BTX = bg.Tx

	go func() {
		time.Sleep(time.Second * 3)
		cmd.Tx <- "i 1 | s fartcoinusdt"
	}()

	root, err := comps.RootFromConfig(&c)
	if err != nil {
		panic(err)
	}

	app := &App{
		S:    state,
		Root: root,
		CMD:  cmd,
		BG:   bg,
		C:    core.NewController(),
	}

	defer rl.CloseWindow()

	rl.SetTargetFPS(c.TargetFPS)

	for !rl.WindowShouldClose() {
		app.Frame()
	}

	return nil
}

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
	"strings"
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

	a.S.WHF = rl.IsWindowHidden()
	a.S.WFF = rl.IsWindowFocused()
	a.S.WRF = rl.IsWindowResized() || a.S.FN == 0
	if a.S.WRF {
		sW, sH := rl.GetScreenWidth(), rl.GetScreenHeight()
		a.S.WS = rl.NewVector2(float32(sW), float32(sH))
	}

	a.S.ThrottlingF = a.S.FN%uint64(a.S.TFPS/3) == 0

	a.C.Event(a.S)
	a.BG.Update(a.S)

	rl.BeginDrawing()
	a.Root.Render(a.S)
	rl.EndDrawing()

	a.S.FN += 1

	a.CMD.Update(a.S)
}

func InitApp(ctx context.Context, c *core.Config) *App {

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
		bybit.WithHttpClient(httpClient),
	)

	br := bybit.NewBroker(client)

	state := core.InitState(c)

	root, err := comps.InitRoot(c, state)
	if err != nil {
		panic(err)
	}

	bg := core.InitBackground(ctx, br)

	cmd := core.InitCMD(ctx, c)
	cmd.BTX = bg.Tx

	controller := core.InitController(c)
	controller.CMDTX = cmd.Tx

	state.BTX = bg.Tx
	state.CMDTX = cmd.Tx

	if len(c.InitCommands) > 0 {
		go func() {
			time.Sleep(200 * time.Millisecond)
			cmd.Tx <- strings.Join(c.InitCommands, "|")
		}()
	}

	return &App{
		Root: root,
		S:    state,
		CMD:  cmd,
		BG:   bg,
		C:    controller,
	}
}

func Run(ctx context.Context, configPath string) {
	b, err := os.ReadFile(configPath)
	if err != nil {
		panic(err)
	}

	var c core.Config
	if err = json.Unmarshal(b, &c); err != nil {
		panic(err)
	}

	rl.SetConfigFlags(rl.FlagWindowResizable)

	rl.InitWindow(c.InitWindow.Width, c.InitWindow.Height, c.InitWindow.Title)
	rl.SetTraceLogLevel(rl.LogAll)

	rl.SetExitKey(0)

	app := InitApp(ctx, &c)
	app.BG.WatchConfig(ctx, configPath)

	go func() {
		<-ctx.Done()
		rl.CloseWindow()
	}()

	for !rl.WindowShouldClose() {
		app.Frame()
	}
}

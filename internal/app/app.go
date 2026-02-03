package app

import (
	"context"
	"encoding/json"
	"nlkli/raytrade/internal/app/components"
	"nlkli/raytrade/internal/app/config"
	"nlkli/raytrade/internal/app/state"
	"os"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type app struct {
	*state.State

	components []components.Component

	ctx    context.Context
	cancel context.CancelFunc
}

func (a *app) render() {
	rl.ClearBackground(a.P.Bg[1])
	a.Update()

	n := 0
	var actions [8]state.Action
	for _, c := range a.components {
		action := c.Render(a.State)
		if action != nil {
			actions[n] = action
			n += 1
		}
	}
	for i := 0; i < n; i++ {
		a.Apply(actions[i])
	}

	a.FN += 1
}

func Run(ctx context.Context, configPath string) error {
	b, err := os.ReadFile(configPath)
	if err != nil {
		return err
	}
	var config config.Config
	if err = json.Unmarshal(b, &config); err != nil {
		return err
	}
	app := &app{
		State: state.InitState(&config),
		components: []components.Component{
			components.NewCommandLine(),
		},
	}

	rl.SetConfigFlags(rl.FlagWindowResizable)
	rl.InitWindow(config.InitWin.Width, config.InitWin.Height, "raytrade")

	defer rl.CloseWindow()

	rl.SetTargetFPS(60)

	for !rl.WindowShouldClose() {
		rl.BeginDrawing()
		app.render()
		rl.EndDrawing()
	}

	return nil
}

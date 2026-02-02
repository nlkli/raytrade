package app

import (
	"encoding/json"
	"os"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	RH  float32 = 16
	FS  float32 = 14
	WPD float32 = 2

	OB_WIDTH  float32 = 200
	TL_HEIGHT float32 = 20
	PB_WIDTH  float32 = 60
)

type Mode int

const (
	Normal Mode = iota
	Input
)

type rect struct {
	p rl.Vector2
	s rl.Vector2
}

func newRect(pX, pY, sW, sH float32) *rect {
	return &rect{
		p: rl.NewVector2(pX, pY),
		s: rl.NewVector2(sW, sH),
	}
}

func (r *rect) toRectangle() rl.Rectangle {
	return rl.NewRectangle(r.p.X, r.p.Y, r.s.X, r.s.Y)
}

func (r *rect) draw(c rl.Color) {
	rl.DrawRectangleV(r.p, r.s, c)
}

func (r *rect) drawOutline(lt float32, c rl.Color) {
	rl.DrawRectangleLinesEx(r.toRectangle(), lt, c)
}

type imputLine struct {
	r *rect
}

func (il *imputLine) draw(c *colors) {
}

type statusLine struct {
	r *rect
}

func (sl *statusLine) draw(c *colors) {
	sl.r.draw(c.bg[0])
}

type chart struct {
	r *rect

	priceBar *priceBar
	timeLine *timeLine
}

func (ch *chart) draw(c *colors) {
	ch.timeLine = &timeLine{
		r: newRect(ch.r.p.X, ch.r.s.Y-TL_HEIGHT, ch.r.s.X, TL_HEIGHT),
	}
	ch.priceBar = &priceBar{
		r: newRect(ch.r.s.X-PB_WIDTH, ch.r.p.Y, PB_WIDTH, ch.r.s.Y-TL_HEIGHT),
	}
	ch.r.draw(c.bg[1])
	ch.priceBar.draw(c)
	ch.timeLine.draw(c)
}

type priceBar struct {
	r *rect
}

func (pb *priceBar) draw(c *colors) {
	pb.r.draw(c.bg[0])
}

type timeLine struct {
	r *rect
}

func (tl *timeLine) draw(c *colors) {
	tl.r.draw(c.bg[0])
	tl.r.drawOutline(1, c.base.red)
}

type orderBook struct {
	r *rect
}

func (ob *orderBook) draw(c *colors) {
	ob.r.draw(c.bg[1])
}

type state struct {
	ws   rl.Vector2
	mode Mode

	il *imputLine
	sl *statusLine
	ch *chart
	ob *orderBook
}

func (s *state) update() {
	sW, sH := rl.GetScreenWidth(), rl.GetScreenHeight()
	s.ws = rl.NewVector2(float32(sW), float32(sH))

	s.il = &imputLine{
		r: newRect(0, s.ws.Y, s.ws.X, RH*2.33),
	}
	s.sl = &statusLine{
		r: newRect(0, s.ws.Y-s.il.r.s.Y, s.ws.X, RH),
	}
	s.ch = &chart{
		r: newRect(0, 0, s.ws.X-OB_WIDTH, s.sl.r.p.Y),
	}
	s.ob = &orderBook{
		r: newRect(s.ch.r.s.X, 0, OB_WIDTH, s.ch.r.s.Y),
	}

}

type app struct {
	c colors
	*state
}

func (a *app) render() {
	a.update()
	// a.s.charP = rl.GetCharPressed()

	rl.ClearBackground(a.c.bg[1])

	a.state.il.draw(&a.c)
	a.state.sl.draw(&a.c)
	a.state.ch.draw(&a.c)
	a.state.ob.draw(&a.c)
}

func Run(configPath string) error {
	const (
		WIN_WIDTH  = 800
		WIN_HEIGHT = 600
	)

	b, err := os.ReadFile(configPath)
	if err != nil {
		return err
	}
	var config config
	if err = json.Unmarshal(b, &config); err != nil {
		return err
	}

	colors, err := colorsFromConfig(&config)
	if err != nil {
		return err
	}
	app := &app{
		c: colors,
		state: &state{
			ws:   rl.NewVector2(float32(WIN_WIDTH), float32(WIN_HEIGHT)),
			mode: Normal,
		},
	}

	rl.SetConfigFlags(rl.FlagWindowResizable)
	rl.InitWindow(800, 600, "raytrade")

	defer rl.CloseWindow()

	rl.SetTargetFPS(60)

	for !rl.WindowShouldClose() {
		rl.BeginDrawing()
		app.render()
		rl.EndDrawing()
	}

	return nil
}

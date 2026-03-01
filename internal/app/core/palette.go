package core

import (
	"strconv"
	"strings"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type Palette struct {
	Name      string
	IsLight   bool
	Bg        [5]rl.Color
	TBg       [5]rl.Color
	Fg        [4]rl.Color
	Sel       SelectionRlColors
	Cur       CursorRlColors
	Base      AnsiRlColors
	Bright    AnsiRlColors
	Dim       AnsiRlColors
	Diff      DiffRlColors
	CodeSel   [2]rl.Color
	Comment   rl.Color
	OverlayBg rl.Color
}

type SelectionRlColors struct {
	Bg rl.Color
	Fg rl.Color
}

type CursorRlColors struct {
	Bg rl.Color
	Fg rl.Color
}

type AnsiRlColors struct {
	Black   rl.Color
	Red     rl.Color
	Green   rl.Color
	Yellow  rl.Color
	Blue    rl.Color
	Magenta rl.Color
	Cyan    rl.Color
	White   rl.Color
	Orange  rl.Color
	Pink    rl.Color
}

type DiffRlColors struct {
	Add    rl.Color
	Delete rl.Color
	Change rl.Color
	Text   rl.Color
}

func PaletteFromConfig(cfg *Config) *Palette {
	parse := func(s string) rl.Color {
		s = strings.TrimPrefix(s, "#")
		if len(s) != 6 {
			return rl.Color{}
		}
		v, err := strconv.ParseUint(s, 16, 32)
		if err != nil {
			return rl.Color{}
		}
		return rl.Color{
			R: uint8(v >> 16),
			G: uint8((v >> 8) & 0xff),
			B: uint8(v & 0xff),
			A: 255,
		}
	}

	c := cfg.Theme.Colors
	var p Palette

	p.Name = cfg.Theme.Name
	p.IsLight = cfg.Theme.IsLight

	for i := range p.Bg {
		p.Bg[i] = parse(c.Background[i])
		p.TBg[i] = p.Bg[i]
		p.TBg[i].A = 0
	}

	for i := range p.Fg {
		p.Fg[i] = parse(c.Foreground[i])
	}

	p.Sel.Bg = parse(c.Selection.Bg)
	p.Sel.Fg = parse(c.Selection.Fg)

	p.Cur.Bg = parse(c.Cursor.Bg)
	p.Cur.Fg = parse(c.Cursor.Fg)

	fillAnsi := func(src AnsiColors, dst *AnsiRlColors) {
		dst.Black = parse(src.Black)
		dst.Red = parse(src.Red)
		dst.Green = parse(src.Green)
		dst.Yellow = parse(src.Yellow)
		dst.Blue = parse(src.Blue)
		dst.Magenta = parse(src.Magenta)
		dst.Cyan = parse(src.Cyan)
		dst.White = parse(src.White)
		dst.Orange = parse(src.Orange)
		dst.Pink = parse(src.Pink)
	}

	fillAnsi(c.Base, &p.Base)
	fillAnsi(c.Bright, &p.Bright)
	fillAnsi(c.Dim, &p.Dim)

	p.Diff.Add = parse(c.Diff.Add)
	p.Diff.Delete = parse(c.Diff.Delete)
	p.Diff.Change = parse(c.Diff.Change)
	p.Diff.Text = parse(c.Diff.Text)

	for i := range p.CodeSel {
		p.CodeSel[i] = parse(c.CodeSelection[i])
	}

	p.Comment = parse(c.Comment)

	p.OverlayBg = p.Bg[1]
	p.OverlayBg.A /= 2

	return &p
}

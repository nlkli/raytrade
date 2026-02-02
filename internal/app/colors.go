package app

import (
	"fmt"
	"strconv"
	"strings"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type colors struct {
	bg      [5]rl.Color
	fg      [4]rl.Color
	sel     selectionRlColors
	cur     cursorRlColors
	base    ansiRlColors
	bright  ansiRlColors
	dim     ansiRlColors
	diff    diffRlColors
	codeSel [2]rl.Color
	comment rl.Color
}

type selectionRlColors struct {
	bg rl.Color
	fg rl.Color
}

type cursorRlColors struct {
	bg rl.Color
	fg rl.Color
}

type ansiRlColors struct {
	black   rl.Color
	red     rl.Color
	green   rl.Color
	yellow  rl.Color
	blue    rl.Color
	magenta rl.Color
	cyan    rl.Color
	white   rl.Color
	orange  rl.Color
	pink    rl.Color
}

type diffRlColors struct {
	add    rl.Color
	delete rl.Color
	change rl.Color
	text   rl.Color
}

func colorsFromConfig(cfg *config) (colors, error) {
	parse := func(s string) (rl.Color, error) {
		s = strings.TrimPrefix(s, "#")
		if len(s) != 6 {
			return rl.Color{}, fmt.Errorf("invalid color: %s", s)
		}
		v, err := strconv.ParseUint(s, 16, 32)
		if err != nil {
			return rl.Color{}, err
		}
		return rl.Color{
			R: uint8(v >> 16),
			G: uint8((v >> 8) & 0xff),
			B: uint8(v & 0xff),
			A: 255,
		}, nil
	}

	cs := cfg.Theme.Colors

	var c colors
	var err error

	for i := range c.bg {
		if c.bg[i], err = parse(cs.Background[i]); err != nil {
			return c, err
		}
	}
	for i := range c.fg {
		if c.fg[i], err = parse(cs.Foreground[i]); err != nil {
			return c, err
		}
	}

	if c.sel.bg, err = parse(cs.Selection.Bg); err != nil {
		return c, err
	}
	if c.sel.fg, err = parse(cs.Selection.Fg); err != nil {
		return c, err
	}

	if c.cur.bg, err = parse(cs.Cursor.Bg); err != nil {
		return c, err
	}
	if c.cur.fg, err = parse(cs.Cursor.Fg); err != nil {
		return c, err
	}

	fillAnsi := func(src ansiColors, dst *ansiRlColors) error {
		var e error
		if dst.black, e = parse(src.Black); e != nil {
			return e
		}
		if dst.red, e = parse(src.Red); e != nil {
			return e
		}
		if dst.green, e = parse(src.Green); e != nil {
			return e
		}
		if dst.yellow, e = parse(src.Yellow); e != nil {
			return e
		}
		if dst.blue, e = parse(src.Blue); e != nil {
			return e
		}
		if dst.magenta, e = parse(src.Magenta); e != nil {
			return e
		}
		if dst.cyan, e = parse(src.Cyan); e != nil {
			return e
		}
		if dst.white, e = parse(src.White); e != nil {
			return e
		}
		if dst.orange, e = parse(src.Orange); e != nil {
			return e
		}
		if dst.pink, e = parse(src.Pink); e != nil {
			return e
		}
		return nil
	}

	if err = fillAnsi(cs.Base, &c.base); err != nil {
		return c, err
	}
	if err = fillAnsi(cs.Bright, &c.bright); err != nil {
		return c, err
	}
	if err = fillAnsi(cs.Dim, &c.dim); err != nil {
		return c, err
	}

	if c.diff.add, err = parse(cs.Diff.Add); err != nil {
		return c, err
	}
	if c.diff.delete, err = parse(cs.Diff.Delete); err != nil {
		return c, err
	}
	if c.diff.change, err = parse(cs.Diff.Change); err != nil {
		return c, err
	}
	if c.diff.text, err = parse(cs.Diff.Text); err != nil {
		return c, err
	}

	for i := range c.codeSel {
		if c.codeSel[i], err = parse(cs.CodeSelection[i]); err != nil {
			return c, err
		}
	}

	if c.comment, err = parse(cs.Comment); err != nil {
		return c, err
	}

	return c, nil
}

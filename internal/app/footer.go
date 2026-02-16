package app

import (
	"fmt"
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type Footer struct {
	*Rect

	sl StatusLine
	cl CommandLine
}

func (f *Footer) Render(s *State) {
	if s.WRF || s.RH_Dirty {
		f.sl.Rect, f.cl.Rect = f.SplitH(s.RH)
	}

	f.sl.Render(s)
	f.cl.Render(s)

	// f.Outline(1, s.P.Base.Pink)
}

type StatusLine struct {
	*Rect

	utS  string
	utTW float32
}

func (sl *StatusLine) Render(s *State) {
	sl.Fill(s.P.Bg[0])

	if s.FN%uint64(s.TFPS/4) == 0 {
		sl.utS = time.Since(s.ST).Truncate(time.Second).String()
		sl.utTW = s.StdMeasureText(sl.utS).X
	}

	if len(sl.utS) > 0 {
		s.StdDrawText(sl.utS, rl.Vector2{X: sl.s.X - sl.utTW, Y: sl.p.Y}, s.P.Fg[3])
	}

	s.StdDrawText(
		fmt.Sprintf("%s/%s", s.StatusLine.Symbol, s.StatusLine.Interval),
		rl.Vector2{X: sl.p.X, Y: sl.p.Y},
		s.P.Fg[3],
	)
}

type CommandLine struct {
	*Rect
}

func (cl *CommandLine) Render(s *State) {
	if s.M == Input {
		rl.DrawRectangleV(
			rl.Vector2{
				X: cl.p.X + s.CommandLine.PromptW + 1,
				Y: cl.p.Y + s.CommandLine.LinesH,
			},
			rl.Vector2{X: s.RH / 2, Y: s.RH},
			s.P.Cur.Bg,
		)
	}

	// if len(s.CommandLine.Lines) > 0 {
	// 	for i, l := range s.CommandLine.Lines {
	// 		height := s.RH * float32(i+1) // TODO limit
	// 		s.StdDrawText(
	// 			l,
	// 			rl.Vector2{X: cl.p.X, Y: cl.parent.p.Y + height},
	// 			s.CommandLine.Color,
	// 		)
	// 	}
	// }

	if len(s.CommandLine.Prompt) > 0 {
		s.StdDrawText(
			s.CommandLine.Prompt,
			rl.Vector2{
				X: cl.p.X,
				Y: cl.p.Y + s.CommandLine.LinesH,
			},
			s.CommandLine.Color,
		)
	}

	// cl.Outline(1, s.P.Base.Pink)
}

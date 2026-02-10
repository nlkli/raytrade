package app

import (
	"context"
	"fmt"
	"nlkli/raytrade/internal/cdl"
	"strconv"
	"strings"
	"sync/atomic"
)

type CMD struct {
	Tx   chan string
	slot atomic.Pointer[CommitFn]
}

func InitCMD(ctx context.Context) *CMD {
	c := &CMD{
		Tx: make(chan string, 32),
	}

	go func() {
		for {
			select {
			case p := <-c.Tx:
				f := c.translate(p)
				c.slot.Store(&f)
			case <-ctx.Done():
				return
			}
		}
	}()

	return c
}

func (c *CMD) Update(s *State) error {
	if ptr := c.slot.Swap(nil); ptr != nil {
		return (*ptr)(s)
	}
	return nil
}

func (c *CMD) translate(prompt string) CommitFn {

	cmdError := func(text string) func(s *State) error {
		return func(s *State) error {
			s.CommandLine.Prompt = text
			s.CommandLine.Color = s.P.Base.Red
			return nil
		}
	}

	parseFloatValue := func(v string, f *float64) (err error) {
		switch rune(v[0]) {
		case '+':
			pf, err := strconv.ParseFloat(v[1:], 64)
			if err != nil {
				return err
			}
			*f += pf
		case '-':
			pf, err := strconv.ParseFloat(v[1:], 64)
			if err != nil {
				return err
			}
			*f -= pf
		default:
			*f, err = strconv.ParseFloat(v, 64)
		}
		return err
	}

	if len(prompt) == 0 {
		return cmdError("empty command")
	}

	var commands []CommitFn

	parts := strings.SplitSeq(prompt, "|")

	for part := range parts {

		part = strings.TrimSpace(part)

		args := strings.Split(part, " ")
		n := len(args)

		var command CommitFn

		switch args[0] {

		case "symbol", "s":

			if n == 1 {
				break
			}

			symbol := strings.ToUpper(args[1])
			command = func(s *State) error {
				s.StatusLine.Symbol = symbol
				return nil
			}

		case "interval", "i":

			if n == 1 {
				break
			}

			interval, err := cdl.IntervalFromString(args[1])

			if err != nil {
				return cmdError(err.Error())
			}

			command = func(s *State) error {
				s.StatusLine.Interval = interval.AsString()
				return nil
			}

		case "reset", "r":

			if n == 1 {
				break
			}

			switch args[1] {

			// case "ut":
			// 	tn := time.Now()

			// 	command = func(s *State) error {
			// 		s.ST = tn
			// 		return nil
			// 	}

			case "scalex", "sx":

				if n == 1 {
					break
				}

				command = func(s *State) error {
					s.Chart.Forced = true
					s.Chart.Scale.X = DEFAULT_SCALE_X
					return nil
				}

			case "scaley", "sy":

				if n == 1 {
					break
				}

				command = func(s *State) error {
					s.Chart.Forced = true
					s.Chart.Scale.Y = DEFAULT_SCALE_Y
					return nil
				}

			case "targetx", "tx":

				if n == 1 {
					break
				}

				command = func(s *State) error {
					s.Chart.Forced = true
					s.Chart.Shift.X = DEFAULT_SHIFT_X
					return nil
				}

			case "targety", "ty":

				if n == 1 {
					break
				}

				command = func(s *State) error {
					s.Chart.Forced = true
					s.Chart.Shift.Y = DEFAULT_SHIFT_Y
					return nil
				}

			case "rowheight", "rh":

				if n == 1 {
					break
				}

				command = func(s *State) error {
					s.RH = DEFAULT_ROW_HEIGHT
					s.Footer.Forced = true
					s.Chart.Forced = true
					return nil
				}

			default:
			}

		case "scalex", "sx":

			if n == 1 {
				break
			}

			command = func(s *State) error {
				f := float64(s.Chart.Scale.X)
				if err := parseFloatValue(args[1], &f); err != nil {
					return cmdError(err.Error())(s)
				}
				s.Chart.Scale.X = float32(f)
				s.Chart.Forced = true
				return nil
			}

		case "scaley", "sy":

			if n == 1 {
				break
			}

			command = func(s *State) error {
				f := float64(s.Chart.Scale.Y)
				if err := parseFloatValue(args[1], &f); err != nil {
					return cmdError(err.Error())(s)
				}
				s.Chart.Scale.Y = float32(f)
				s.Chart.Forced = true
				return nil
			}

		case "targetx", "tx":

			if n == 1 {
				break
			}

			command = func(s *State) error {
				f := float64(s.Chart.Shift.X)
				if err := parseFloatValue(args[1], &f); err != nil {
					return cmdError(err.Error())(s)
				}
				s.Chart.Shift.X = float32(f)
				s.Chart.Forced = true
				return nil
			}

		case "targety", "ty":

			if n == 1 {
				break
			}

			command = func(s *State) error {
				f := float64(s.Chart.Shift.Y)
				if err := parseFloatValue(args[1], &f); err != nil {
					return cmdError(err.Error())(s)
				}
				s.Chart.Shift.Y = float32(f)
				s.Chart.Forced = true
				return nil
			}

		case "rowheight", "rh":

			if n == 1 {
				break
			}

			command = func(s *State) error {
				f := float64(s.RH)
				if err := parseFloatValue(args[1], &f); err != nil {
					return cmdError(err.Error())(s)
				}
				s.RH = float32(f)
				s.Footer.Forced = true
				s.Chart.Forced = true
				return nil
			}

		default:
		}

		if command != nil {
			commands = append(commands, command)
		}
	}

	if len(commands) == 1 {
		return commands[0]
	}
	if len(commands) > 1 {
		return func(s *State) error {
			for _, c := range commands {
				c(s)
			}
			return nil
		}
	}

	return cmdError(fmt.Sprintf("unknown command: %s", prompt))
}

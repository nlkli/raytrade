package app

import (
	"context"
	"fmt"
	"nlkli/raytrade/internal/broker"
	"nlkli/raytrade/internal/cdl"
	"strconv"
	"strings"
)

type Task any

type CommandPromptT struct {
	Prompt string
}

type Command func(*State) error

type Background struct {
	Tx chan Task
	Rx chan Command

	broker broker.Broker
}

func NewBackground(ctx context.Context, br broker.Broker) *Background {
	w := &Background{
		Tx: make(chan Task, 32),
		Rx: make(chan Command, 32),

		broker: br,
	}

	go func() {
		for {
			select {
			case t := <-w.Tx:
				switch t := t.(type) {
				case CommandPromptT:
					w.Rx <- cmd(t.Prompt)
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	return w
}

func cmd(prompt string) Command {
	if len(prompt) == 0 {
		return cmdError("empty command")
	}

	var commands []Command

	parts := strings.SplitSeq(prompt, "|")

	for part := range parts {

		part = strings.TrimSpace(part)

		args := strings.Split(part, " ")
		n := len(args)

		var command Command

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
					return nil
				}

			default:
			}

		case "scalex", "sx":

			if n == 1 {
				break
			}

			command = func(s *State) error {
				s.Chart.Forced = true
				f := float64(s.Chart.Scale.X)
				parseFloatValue(args[1], &f)
				s.Chart.Scale.X = float32(f)
				return nil
			}

		case "scaley", "sy":

			if n == 1 {
				break
			}

			command = func(s *State) error {
				s.Chart.Forced = true
				f := float64(s.Chart.Scale.Y)
				parseFloatValue(args[1], &f)
				s.Chart.Scale.Y = float32(f)
				return nil
			}

		case "targetx", "tx":

			if n == 1 {
				break
			}

			command = func(s *State) error {
				s.Chart.Forced = true
				f := float64(s.Chart.Shift.X)
				parseFloatValue(args[1], &f)
				s.Chart.Shift.X = float32(f)
				return nil
			}

		case "targety", "ty":

			if n == 1 {
				break
			}

			command = func(s *State) error {
				s.Chart.Forced = true
				f := float64(s.Chart.Shift.Y)
				parseFloatValue(args[1], &f)
				s.Chart.Shift.Y = float32(f)
				return nil
			}

		case "rowheight", "rh":

			if n == 1 {
				break
			}

			command = func(s *State) error {
				f := float64(s.RH)
				parseFloatValue(args[1], &f)
				s.RH = float32(f)
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
				if c != nil {
					c(s)
				}
			}
			return nil
		}
	}

	return cmdError(fmt.Sprintf("unknown command: %s", prompt))
}

func cmdError(text string) func(s *State) error {
	return func(s *State) error {
		s.CommandLine.Prompt = text
		s.CommandLine.Color = s.P.Base.Red
		return nil
	}
}

func parseFloatValue(v string, f *float64) (err error) {
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

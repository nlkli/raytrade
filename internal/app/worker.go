package app

import (
	"context"
	"fmt"
	"nlkli/raytrade/internal/broker"
	"nlkli/raytrade/internal/cdl"
	"strconv"
	"strings"
	"time"
)

type Task any

type CommandPromptT struct {
	Prompt string
}

type Command func(*State) error

type Worker struct {
	Tx chan Task
	Rx chan Command

	broker broker.Broker
}

func NewWorker(ctx context.Context, br broker.Broker) *Worker {
	w := &Worker{
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
				s.StatusLine.Interval = interval
				return nil
			}

		case "reset", "r":

			if n == 1 {
				break
			}

			switch args[1] {

			case "ut":
				tn := time.Now()
				return func(s *State) error {
					s.ST = tn
					return nil
				}

			default:
			}

		case "scalex", "scx":

			if n == 1 {
				break
			}

			f, err := strconv.ParseFloat(args[1], 64)
			if err != nil {
				return cmdError(err.Error())
			}

			command = func(s *State) error {
				s.Chart.scale.X = float32(f)
				return nil
			}

		case "scaley", "scy":

			if n == 1 {
				break
			}

			f, err := strconv.ParseFloat(args[1], 64)
			if err != nil {
				return cmdError(err.Error())
			}

			command = func(s *State) error {
				s.Chart.scale.Y = float32(f)
				return nil
			}

		case "shiftx", "shx":

			if n == 1 {
				break
			}

			f, err := strconv.ParseFloat(args[1], 64)
			if err != nil {
				return cmdError(err.Error())
			}

			command = func(s *State) error {
				s.Chart.shift.X = float32(f)
				return nil
			}

		case "shifty", "shy":

			if n == 1 {
				break
			}

			f, err := strconv.ParseFloat(args[1], 64)
			if err != nil {
				return cmdError(err.Error())
			}

			command = func(s *State) error {
				s.Chart.shift.Y = float32(f)
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

package app2

import (
	"context"
	"fmt"
	"strings"
	"time"
)

type Worker struct {
	Tx chan Task
	Rx chan func(*State) error
}

func NewWorker(ctx context.Context) *Worker {
	w := &Worker{
		Tx: make(chan Task, 32),
		Rx: make(chan func(*State) error, 32),
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

func cmd(prompt string) func(*State) error {
	if len(prompt) == 0 {
		return func(s *State) error {
			s.CommandLine.Prompt = "Empty command"
			s.CommandLine.Color = s.P.Base.Red
			return nil
		}
	}

	prompt = strings.TrimSpace(prompt)

	args := strings.Split(prompt, " ")
	n := len(args)

	switch args[0] {

	case "reset":

		if n == 2 {
			switch args[1] {

			case "ut":
				tn := time.Now()
				return func(s *State) error {
					s.ST = tn
					return nil
				}

			default:
			}
		}

	default:
	}

	return func(s *State) error {
		s.CommandLine.Prompt = fmt.Sprintf("Unknown command: %s", prompt)
		s.CommandLine.Color = s.P.Base.Red
		return nil
	}
}

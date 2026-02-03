package app2

import (
	"context"
	"fmt"
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
				case CommandPromptTask:
					w.Rx<-cmd(t.Prompt)
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	return w
}

func cmd(prompt string) func(*State) error {
	return func(s *State) error {
		s.CommandLine.Prompt = fmt.Sprintf("Unknown command: %s", prompt)
		s.CommandLine.Color = s.P.Base.Red
		return nil
	}
}

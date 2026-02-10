package app

import (
	"context"
	"nlkli/raytrade/internal/broker"
)

type Task any

type Background struct {
	Tx chan Task
	Rx chan CommitFn

	broker broker.Broker
}

func InitBackground(ctx context.Context, br broker.Broker) *Background {
	w := &Background{
		Tx: make(chan Task, 32),
		Rx: make(chan CommitFn, 32),

		broker: br,
	}

	// go func() {
	// 	for {
	// 		select {
	// 		case _ := <-w.Tx:
	// 		case <-ctx.Done():
	// 			return
	// 		}
	// 	}
	// }()

	return w
}

func (b *Background) Update(s *State) {

}

package utils

import "sync"

type Future[T any] struct {
	ch         chan T
	once       sync.Once
	onComplete func()
}

func NewFuture[T any](onComplete func()) *Future[T] {
	return &Future[T]{
		ch:         make(chan T, 1),
		onComplete: onComplete,
	}
}

func (f *Future[T]) Complete(res T) {
	f.once.Do(func() {
		f.ch <- res
		if f.onComplete != nil {
			f.onComplete()
		}
		close(f.ch)
	})
}

func (f *Future[T]) Await() <-chan T {
	return f.ch
}

package utils

import (
	"sync"
)

func MergeSubscriptions[T any](buffer int, subs ...*Subscription[T]) *Subscription[T] {
	if len(subs) == 0 {
		return nil
	}
	if len(subs) == 1 {
		return subs[0]
	}

	ch := make(chan T, buffer)
	done := make(chan struct{})

	var wg sync.WaitGroup

	s := &Subscription[T]{
		ch: ch,
		onStop: func() {
			close(done)

			for _, sub := range subs {
				if sub != nil {
					sub.Stop()
				}
			}

			wg.Wait()
			close(ch)
		},
	}

	for _, sub := range subs {
		if sub == nil {
			continue
		}

		wg.Add(1)
		go func() {
			defer s.Stop()
			defer wg.Done()

			for {
				select {
				case v, ok := <-sub.ch:
					if !ok {
						return
					}
					select {
					case ch <- v:
					case <-done:
						return
					default:
					}
				case <-done:
					return
				}
			}
		}()
	}

	return s
}

type Subscription[T any] struct {
	ch     chan T
	once   sync.Once
	onStop func()
}

func (s *Subscription[T]) C() <-chan T {
	return s.ch
}

func (s *Subscription[T]) Stop() {
	s.once.Do(s.onStop)
}

type Broadcaster[T any] struct {
	mu   sync.RWMutex
	subs map[*Subscription[T]]struct{}
}

func NewBroadcaster[T any]() *Broadcaster[T] {
	return &Broadcaster[T]{
		subs: make(map[*Subscription[T]]struct{}),
	}
}

func (b *Broadcaster[T]) Subscribe(buffer int, onStop func(*Broadcaster[T])) *Subscription[T] {
	ch := make(chan T, buffer)

	sub := &Subscription[T]{
		ch: ch,
	}

	sub.onStop = func() {
		b.mu.Lock()
		delete(b.subs, sub)
		b.mu.Unlock()

		if onStop != nil {
			onStop(b)
		}

		close(ch)
	}

	b.mu.Lock()
	b.subs[sub] = struct{}{}
	b.mu.Unlock()

	return sub
}

func (b *Broadcaster[T]) Publish(event T) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	for sub := range b.subs {
		select {
		case sub.ch <- event:
		default:
		}
	}
}

func (b *Broadcaster[T]) Subs() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.subs)
}

func (b *Broadcaster[T]) IsEmpty() bool {
	return b.Subs() == 0
}

func (b *Broadcaster[T]) Clear() {
	b.mu.Lock()
	snapshot := make([]*Subscription[T], 0, len(b.subs))
	for sub := range b.subs {
		snapshot = append(snapshot, sub)
	}
	b.mu.Unlock()

	for _, sub := range snapshot {
		if sub != nil {
			sub.Stop()
		}
	}
}

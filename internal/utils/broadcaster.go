package utils

import "sync"

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

	for _, sub := range subs {
		if sub == nil {
			continue
		}

		wg.Add(1)
		go func(s *Subscription[T]) {
			defer wg.Done()
			for {
				select {
				case v, ok := <-s.C():
					if !ok {
						return
					}
					select {
					case ch <- v:
					case <-done:
						return
					}
				case <-done:
					return
				}
			}
		}(sub)
	}

	return &Subscription[T]{
		ch: ch,
		stop: func() {
			close(done)

			for _, sub := range subs {
				if sub != nil {
					sub.Unsubscribe()
				}
			}

			wg.Wait()
			close(ch)
		},
	}
}

type Subscription[T any] struct {
	ch   chan T
	once sync.Once
	stop func()
}

func (s *Subscription[T]) C() <-chan T {
	return s.ch
}

func (s *Subscription[T]) Unsubscribe() {
	if s.stop != nil {
		s.once.Do(s.stop)
	}
}

type Broadcaster[T any] struct {
	subscribers map[chan<- T]struct{}
	mu          sync.RWMutex
}

func NewBroadcaster[T any]() *Broadcaster[T] {
	return &Broadcaster[T]{
		subscribers: make(map[chan<- T]struct{}),
	}
}

func (b *Broadcaster[T]) Subscribe(buffer int, onUnsubscribe func(*Broadcaster[T])) *Subscription[T] {
	ch := make(chan T, buffer)

	b.mu.Lock()
	b.subscribers[ch] = struct{}{}
	b.mu.Unlock()

	sub := &Subscription[T]{
		ch: ch,
		stop: func() {
			b.mu.Lock()
			delete(b.subscribers, ch)
			b.mu.Unlock()

			if onUnsubscribe != nil {
				onUnsubscribe(b)
			}

			close(ch)
		},
	}

	return sub
}

func (b *Broadcaster[T]) Publish(event T) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	for ch := range b.subscribers {
		select {
		case ch <- event:
		default:
		}
	}
}

func (b *Broadcaster[T]) Subs() int {
	b.mu.RLock()
	defer b.mu.RUnlock()

	return len(b.subscribers)
}

func (b *Broadcaster[T]) IsEmpty() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()

	return len(b.subscribers) == 0
}


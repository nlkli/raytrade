package bybit

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"nlkli/raytrade/internal/broker/bybit/models"
	"nlkli/raytrade/internal/utils"
	"nlkli/raytrade/internal/ws"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

var globReqID atomic.Uint64

func newStreamOpRequest(
	op models.StreamOperation,
	topics []string,
) (string, []byte, error) {
	reqID := strconv.FormatUint(globReqID.Add(1), 10)

	req := models.StreamOperationRequest{
		ReqID: reqID,
		Op:    op,
		Args:  make([]any, len(topics)),
	}
	for i, t := range topics {
		req.Args[i] = t
	}

	b, err := json.Marshal(req)
	return reqID, b, err
}

type Stream struct {
	conn *ws.Conn

	topics map[string]*utils.Broadcaster[*models.StreamData]
	ops    map[string]*utils.Future[error]
	mu     sync.RWMutex
	wg     sync.WaitGroup

	subscribeBarrier sync.Mutex

	ctx    context.Context
	cancel context.CancelFunc
}

func NewStream(
	ctx context.Context,
	url string,
	onOpen func(*ws.Conn) error,
) *Stream {
	s := &Stream{
		topics: make(map[string]*utils.Broadcaster[*models.StreamData]),
		ops:    make(map[string]*utils.Future[error]),
	}

	s.ctx, s.cancel = context.WithCancel(ctx)

	s.conn = ws.NewConn(s.ctx, url, func(conn *ws.Conn) error {
		if onOpen != nil {
			if err := onOpen(conn); err != nil {
				return err
			}
		}

		topics := s.Topics()
		if len(topics) == 0 {
			return nil
		}

		_, b, err := newStreamOpRequest(
			models.StreamOperationSubscribe,
			topics,
		)
		if err != nil {
			return err
		}

		return conn.Send(b)
	})

	s.wg.Go(s.run)

	return s
}

func (s *Stream) Topics() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	out := make([]string, 0, len(s.topics))
	for t := range s.topics {
		out = append(out, t)
	}
	return out
}

func (s *Stream) Subscribe(topics []string, buffer int) (*utils.Subscription[*models.StreamData], error) {
	if len(topics) == 0 {
		return nil, fmt.Errorf("empty topics")
	}

	s.subscribeBarrier.Lock()
	defer s.subscribeBarrier.Unlock()

	set := make(map[string]struct{}, len(topics))
	for _, t := range s.Topics() {
		set[t] = struct{}{}
	}

	newTopics := make([]string, 0)
	for _, t := range topics {
		if _, ok := set[t]; !ok {
			newTopics = append(newTopics, t)
		}
	}

	if len(newTopics) > 0 {
		reqID, b, err := newStreamOpRequest(models.StreamOperationSubscribe, newTopics)
		if err != nil {
			return nil, err
		}

		fut := utils.NewFuture[error](func() {
			s.mu.Lock()
			delete(s.ops, reqID)
			s.mu.Unlock()
		})

		defer fut.Complete(nil)

		s.mu.Lock()
		s.ops[reqID] = fut
		s.mu.Unlock()

		if err := s.conn.Send(b); err != nil {
			return nil, err
		}

		select {
		case err := <-fut.Await():
			if err != nil {
				return nil, err
			}
		case <-time.After(5 * time.Second):
			return nil, errors.New("subscribe timeout")
		case <-s.ctx.Done():
			return nil, errors.New("stream closed")
		}
	}

	subs := make([]*utils.Subscription[*models.StreamData], 0)

	s.mu.Lock()
	for _, t := range topics {
		if s.topics[t] == nil {
			s.topics[t] = utils.NewBroadcaster[*models.StreamData]()
		}
		sub := s.topics[t].Subscribe(buffer, func(b *utils.Broadcaster[*models.StreamData]) {
			if b.IsEmpty() {
				s.subscribeBarrier.Lock()
				defer s.subscribeBarrier.Unlock()

				s.mu.Lock()
				delete(s.topics, t)
				s.mu.Unlock()

				_, b, err := newStreamOpRequest(models.StreamOperationUnsubscribe, []string{t})
				if err != nil {
					return
				}

				s.conn.Send(b)
			}
		})
		subs = append(subs, sub)
	}
	s.mu.Unlock()

	return utils.MergeSubscriptions(buffer, subs...), nil
}

func (s *Stream) run() {
	successKey := []byte(`"success":`)

	for {
		b, err := s.conn.Recv()
		if err != nil {
			if s.ctx.Err() != nil {
				return
			}
			continue
		}

		if !bytes.Contains(b[:min(len(b), 128)], successKey) {
			var data models.StreamData
			if json.Unmarshal(b, &data) == nil {
				s.mu.RLock()
				if bc, ok := s.topics[data.Topic]; ok {
					bc.Publish(&data)
				}
				s.mu.RUnlock()
			}
			continue
		}

		var res models.StreamOperationResult
		if json.Unmarshal(b, &res) == nil {
			s.mu.RLock()
			if fu, ok := s.ops[res.ReqID]; ok {
				if res.Success {
					fu.Complete(nil)
				} else {
					fu.Complete(errors.New(res.RetMsg))
				}
			}
			s.mu.RUnlock()
		}
	}
}

func (s *Stream) Close() {
	s.subscribeBarrier.Lock()
	defer s.subscribeBarrier.Unlock()

	s.cancel()
	s.wg.Wait()
}

// func NewStream(
// 	ctx context.Context,
// 	url string,
// 	onOpen func(*ws.Conn) error,
// ) *Stream {
// 	s := &Stream{
// 		topics: make(map[string][]*subscriber),
// 		ops:    make(map[string]chan error),
// 	}
//
// 	s.ctx, s.cancel = context.WithCancel(ctx)
//
// 	s.conn = ws.NewConn(ctx, url, func(conn *ws.Conn) error {
// 		if onOpen != nil {
// 			if err := onOpen(conn); err != nil {
// 				return err
// 			}
// 		}
//
// 		topics := s.Topics()
// 		if len(topics) == 0 {
// 			return nil
// 		}
//
// 		_, b, err := newStreamOpRequest(
// 			models.StreamOperationSubscribe,
// 			topics,
// 		)
// 		if err != nil {
// 			return err
// 		}
//
// 		return conn.Send(b)
// 	})
//
// 	s.wg.Go(s.run)
//
// 	return s
// }
//
// func (s *Stream) Subscribe(
// 	topics []string,
// 	buf int,
// ) (<-chan *models.StreamData, error) {
// 	if len(topics) == 0 {
// 		return nil, fmt.Errorf("empty topics")
// 	}
//
// 	reqID, b, err := newStreamOpRequest(
// 		models.StreamOperationSubscribe,
// 		topics,
// 	)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	resCh := make(chan error, 1)
//
// 	s.mu.Lock()
// 	s.ops[reqID] = resCh
// 	s.mu.Unlock()
//
// 	defer func() {
// 		s.mu.Lock()
// 		delete(s.ops, reqID)
// 		s.mu.Unlock()
// 		close(resCh)
// 	}()
//
// 	if err := s.conn.Send(b); err != nil {
// 		return nil, err
// 	}
//
// 	select {
// 	case err, ok := <-resCh:
// 		if !ok {
// 			return nil, errors.New("stream closed")
// 		}
// 		if err != nil {
// 			return nil, err
// 		}
// 	case <-s.ctx.Done():
// 		return nil, s.ctx.Err()
// 	}
//
// 	sub := &subscriber{
// 		ch: make(chan *models.StreamData, buf),
// 	}
//
// 	s.mu.Lock()
// 	for _, t := range topics {
// 		s.topics[t] = append(s.topics[t], sub)
// 	}
// 	delete(s.ops, reqID)
// 	s.mu.Unlock()
//
// 	return sub.ch, nil
// }
//
// func (s *Stream) Unsubscribe(ch chan *models.StreamData) error {
// 	if ch == nil {
// 		return fmt.Errorf("nil channel")
// 	}
//
// 	var unsubscribe []string
//
// 	s.mu.Lock()
// 	for topic, subs := range s.topics {
// 		n := subs[:0]
// 		for _, sub := range subs {
// 			if sub.ch != ch {
// 				n = append(n, sub)
// 			}
// 		}
//
// 		if len(n) == 0 {
// 			unsubscribe = append(unsubscribe, topic)
// 			delete(s.topics, topic)
// 		} else {
// 			s.topics[topic] = n
// 		}
// 	}
//
// 	close(ch)
//
// 	s.mu.Unlock()
//
// 	if len(unsubscribe) > 0 {
// 		if _, b, err := newStreamOpRequest(
// 			models.StreamOperationUnsubscribe,
// 			unsubscribe,
// 		); err == nil {
// 			_ = s.conn.Send(b)
// 		}
// 	}
//
// 	return nil
// }
//
// func (s *Stream) Topics() []string {
// 	s.mu.RLock()
// 	defer s.mu.RUnlock()
//
// 	out := make([]string, 0, len(s.topics))
// 	for t := range s.topics {
// 		out = append(out, t)
// 	}
// 	return out
// }
//
// func (s *Stream) broadcast(data *models.StreamData) {
// 	s.mu.RLock()
// 	subs := append([]*subscriber(nil), s.topics[data.Topic]...)
// 	s.mu.RUnlock()
//
// 	for _, sub := range subs {
// 		select {
// 		case sub.ch <- data:
// 		default:
// 		}
// 	}
// }
//
// func (s *Stream) run() {
// 	successKey := []byte(`"success":`)
//
// 	for {
// 		b, err := s.conn.Recv()
// 		if err != nil {
// 			if s.ctx.Err() != nil {
// 				return
// 			}
// 			continue
// 		}
//
// 		if !bytes.Contains(b[:min(len(b), 128)], successKey) {
// 			var data *models.StreamData
// 			if json.Unmarshal(b, &data) == nil {
// 				s.broadcast(data)
// 			}
// 			continue
// 		}
//
// 		var res models.StreamOperationResult
// 		if json.Unmarshal(b, &res) == nil {
// 			s.mu.RLock()
// 			ch := s.ops[res.ReqID]
// 			s.mu.RUnlock()
//
// 			if ch != nil {
// 				if res.Success {
// 					ch <- nil
// 				} else {
// 					ch <- errors.New(res.RetMsg)
// 				}
// 			}
// 		}
// 	}
// }
//
// func (s *Stream) Close() {
// 	s.cancel()
// 	s.conn.Close()
//
// 	s.wg.Wait()
//
// 	s.mu.Lock()
// 	defer s.mu.Unlock()
//
// 	for _, subs := range s.topics {
// 		for _, sub := range subs {
// 			close(sub.ch)
// 		}
// 	}
// 	s.topics = nil
//
// 	for _, ch := range s.ops {
// 		close(ch)
// 	}
// 	s.ops = nil
// }

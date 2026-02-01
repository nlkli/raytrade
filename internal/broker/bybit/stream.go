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

func nextReqID() string {
	return strconv.FormatUint(globReqID.Add(1), 10)
}

func newStreamOpRequest(
	op models.StreamOp,
	args []string,
) (string, []byte, error) {
	reqID := nextReqID()
	req := models.StreamOpRequest{
		ReqID: reqID,
		Op:    op,
		Args:  make([]any, len(args)),
	}
	for i, a := range args {
		req.Args[i] = a
	}

	b, err := json.Marshal(req)
	return reqID, b, err
}

type Stream struct {
	conn *ws.Conn

	topics map[string]*utils.Broadcaster[*models.StreamData]
	ops    map[string]*utils.Future[error]

	mu         sync.RWMutex
	wg         sync.WaitGroup
	subBarrier sync.Mutex

	ctx    context.Context
	cancel context.CancelFunc
}

func NewStream(
	ctx context.Context,
	url string,
	onOpen func(*ws.Conn) error,
) *Stream {
	ctx, cancel := context.WithCancel(ctx)

	s := &Stream{
		topics: make(map[string]*utils.Broadcaster[*models.StreamData]),
		ops:    make(map[string]*utils.Future[error]),

		ctx:    ctx,
		cancel: cancel,
	}

	s.conn = ws.NewConn(ctx, url, func(conn *ws.Conn) error {
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
			models.StreamOpSubscribe,
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
	if err := s.ctx.Err(); err != nil {
		return nil, err
	}

	s.subBarrier.Lock()
	defer s.subBarrier.Unlock()

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
		reqID, b, err := newStreamOpRequest(models.StreamOpSubscribe, newTopics)
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
				s.subBarrier.Lock()
				defer s.subBarrier.Unlock()

				s.mu.Lock()
				delete(s.topics, t)
				s.mu.Unlock()

				_, b, err := newStreamOpRequest(models.StreamOpUnsubscribe, []string{t})
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
				continue
			}
		}

		var res models.StreamOperationResult
		if json.Unmarshal(b, &res) == nil {
			s.mu.RLock()
			fu := s.ops[res.ReqID]
			s.mu.RUnlock()

			if fu != nil {
				if res.Success {
					fu.Complete(nil)
					continue
				}
				fu.Complete(errors.New(res.RetMsg))
			}
		}
	}
}

func (s *Stream) Close() {
	s.subBarrier.Lock()
	s.cancel()
	s.wg.Wait()
	s.subBarrier.Unlock()

	for _, b := range s.topics {
		b.Clear()
	}
}

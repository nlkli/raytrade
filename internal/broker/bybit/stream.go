package bybit

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"nlkli/raytrade/internal/broker/bybit/models"
	"nlkli/raytrade/internal/ws"
	"strconv"
	"sync"
	"sync/atomic"
)

var globReqID atomic.Uint64

func newStreamOperationRequest(op models.StreamOperation, topics []string) (string, []byte, error) {
	args := make([]any, len(topics))
	for i, t := range topics {
		args[i] = t
	}
	reqID := strconv.FormatUint(globReqID.Add(1), 10)
	req := models.StreamOperationRequest{
		ReqID: reqID,
		Op:    op,
		Args:  args,
	}
	b, err := json.Marshal(req)
	return reqID, b, err
}

type Stream struct {
	conn *ws.Conn

	subTopics   map[string]map[chan<- models.StreamData]struct{}
	subTopicsMu sync.RWMutex

	opResults   map[string]chan<- error
	opResultsMu sync.Mutex

	ctx    context.Context
	cancel context.CancelFunc
}

func NewStream(ctx context.Context, url string, onOpen func(*ws.Conn) error) *Stream {
	s := &Stream{
		subTopics: make(map[string]map[chan<- models.StreamData]struct{}),
		opResults: make(map[string]chan<- error),
	}

	s.conn = ws.NewConn(ctx, url, func(conn *ws.Conn) error {
		if onOpen != nil {
			err := onOpen(conn)
			if err != nil {
				return err
			}
		}
		topics := s.GetTopics()
		if len(topics) > 0 {
			_, b, err := newStreamOperationRequest(models.StreamOperationSubscribe, s.GetTopics())
			if err != nil {
				return err
			}
			return conn.Send(b)
		}
		return nil
	})

	s.ctx, s.cancel = context.WithCancel(ctx)
	go s.run()

	return s
}

func (s *Stream) Subscribe(topics []string, buff int) (<-chan models.StreamData, error) {
	if len(topics) == 0 {
		return nil, fmt.Errorf("empty topics")
	}

	reqID, b, err := newStreamOperationRequest(models.StreamOperationSubscribe, topics)
	if err != nil {
		return nil, err
	}

	subResultCh := make(chan error, 1)
	defer func() {
		s.opResultsMu.Lock()
		delete(s.opResults, reqID)
		s.opResultsMu.Unlock()
		close(subResultCh)
	}()

	s.opResultsMu.Lock()
	s.opResults[reqID] = subResultCh
	s.opResultsMu.Unlock()

	err = s.conn.Send(b)
	if err != nil {
		return nil, err
	}

	err = <-subResultCh
	if err != nil {
		return nil, err
	}

	ch := make(chan models.StreamData, buff)

	s.subTopicsMu.Lock()
	for _, t := range topics {
		if s.subTopics[t] == nil {
			s.subTopics[t] = make(map[chan<- models.StreamData]struct{})
		}
		s.subTopics[t][ch] = struct{}{}
	}
	s.subTopicsMu.Unlock()

	return ch, nil
}

func (s *Stream) Unsubscribe(ch chan models.StreamData) error {
	if ch == nil {
		return fmt.Errorf("nil chan")
	}

	defer close(ch)
	unsubscribeTopics := make([]string, 0)

	s.subTopicsMu.Lock()
	for t, subs := range s.subTopics {
		delete(subs, ch)
		if len(subs) == 0 {
			unsubscribeTopics = append(unsubscribeTopics, t)
			delete(s.subTopics, t)
		}
	}
	s.subTopicsMu.Unlock()

	if len(unsubscribeTopics) > 0 {
		_, b, err := newStreamOperationRequest(models.StreamOperationUnsubscribe, unsubscribeTopics)
		if err == nil {
			s.conn.Send(b)
		}
	}
	return nil
}

func (s *Stream) GetTopics() []string {
	topics := make([]string, 0, len(s.subTopics))
	s.subTopicsMu.RLock()
	for k := range s.subTopics {
		topics = append(topics, k)
	}
	s.subTopicsMu.RUnlock()
	return topics
}

func (s *Stream) broadcast(data models.StreamData) {
	s.subTopicsMu.RLock()
	defer s.subTopicsMu.RUnlock()

	for ch := range s.subTopics[data.Topic] {
		select {
		case ch <- data:
		default:
		}
	}
}

func (s *Stream) run() {
	for {
		b, err := s.conn.Recv()
		if err != nil {
			select {
			case <-s.ctx.Done():
				return
			default:
				continue
			}
		}

		var streamEnvelopeMsg models.StreamEnvelopeMessage
		if err := json.Unmarshal(b, &streamEnvelopeMsg); err == nil {

			if streamEnvelopeMsg.Success == nil {
				var data models.StreamData
				if err := json.Unmarshal(b, &data); err == nil {
					s.broadcast(data)
				}
				continue
			}

			var result models.StreamOperationResult
			if err := json.Unmarshal(b, &result); err == nil {
				s.opResultsMu.Lock()
				if ch, ok := s.opResults[result.ReqID]; ok {
					if result.Success {
						ch <- nil
					} else {
						ch <- errors.New(result.RetMsg)
					}
				}
				s.opResultsMu.Unlock()
			}
		}
	}
}

func (s *Stream) Close() {
	s.subTopicsMu.Lock()
	for k, v := range s.subTopics {
		for ch := range v {
			close(ch)
		}
		delete(s.subTopics, k)
	}
	s.subTopicsMu.Unlock()

	s.opResultsMu.Lock()
	for k, ch := range s.opResults {
		close(ch)
		delete(s.opResults, k)
	}
	s.opResultsMu.Unlock()

	s.cancel()
	s.conn.Close()
}

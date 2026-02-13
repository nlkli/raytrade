package bybit

import (
	"encoding/json"
	"errors"
	"nlkli/raytrade/internal/broker/bybit/models"
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

type StreamV2 struct {
	tx         chan<- []byte
	subs       sync.Map
	mu         sync.Mutex
	opResultCh chan models.StreamOpResult
}

func NewStreamV2(
	url string,
	tx chan []byte,
	onConnectedFn ws.OnConnectedFn,
	opts ...ws.PolicyOption,
) *StreamV2 {

	s := new(StreamV2)

	s.opResultCh = make(chan models.StreamOpResult, 96)
	s.tx = tx

	rx := ws.NewConnV2(
		url,
		nil,
		tx,
		0,
		ws.NewPolicy(
			func(sendCh chan<- []byte) error {
				if onConnectedFn != nil {
					if err := onConnectedFn(sendCh); err != nil {
						return nil
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

				sendCh <- b

				return nil
			},
			func(mt int, b []byte) (any, bool) {
				if mt != 1 {
					return nil, true
				}

				var data models.StreamData
				json.Unmarshal(b, &data)

				if len(data.Topic) == 0 {
					var opResult models.StreamOpResult
					if err := json.Unmarshal(b, &opResult); err == nil {
						select {
						case s.opResultCh <- opResult:
						default:
						}
					}

					return nil, true
				}

				return &data, false
			},
			opts...,
		),
	)

	go func() {
		for t := range rx {

			d := t.(*models.StreamData)
			v, ok := s.subs.Load(d.Topic)

			if ok {
				subCh := v.(chan *models.StreamData)
				select {
				case subCh <- d:
				default:
				}
			}
		}

		s.mu.Lock() // close

		s.subs.Range(func(key, v any) bool {
			subCh := v.(chan *models.StreamData)
			close(subCh)

			return true
		})
	}()

	return s
}

func (s *StreamV2) Topics() []string {
	var topics []string

	s.subs.Range(func(key, _ any) bool {
		topics = append(topics, key.(string))
		return true
	})

	return topics
}

func (s *StreamV2) Subscribe(topic string) (chan *models.StreamData, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	reqID, b, err := newStreamOpRequest(
		models.StreamOpSubscribe,
		[]string{topic},
	)
	if err != nil {
		return nil, err
	}

	if _, ok := s.subs.Load(topic); ok {
		return nil, errors.New("already subscribed")
	}

	subCh := make(chan *models.StreamData)
	s.subs.Store(topic, subCh)

	s.tx <- b

	timeout := time.After(time.Second * 110)

	ticker := time.NewTicker(time.Second * 15)
	defer ticker.Stop()

	for {
		select {
		case opResult := <-s.opResultCh:
			if opResult.ReqID == reqID {
				if !opResult.Success {
					return nil, errors.New(opResult.RetMsg)
				}
				return subCh, nil
			}
		case <-subCh:
			return subCh, nil
		case <-ticker.C:
			s.tx <- b
		case <-timeout:
			return nil, errors.New("timeout")
		}
	}
}

func (s *StreamV2) Unsubscribe(topic string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	reqID, b, err := newStreamOpRequest(
		models.StreamOpUnsubscribe,
		[]string{topic},
	)
	if err != nil {
		return err
	}

	s.tx <- b

	timeout := time.After(time.Second * 110)

	ticker := time.NewTicker(time.Second * 15)
	defer ticker.Stop()

	v, ok := s.subs.LoadAndDelete(topic)
	if ok {
		subCh := v.(chan *models.StreamData)
		close(subCh)
	}

	for {
		select {
		case opResult := <-s.opResultCh:
			if opResult.ReqID == reqID {
				if !opResult.Success {
					return errors.New(opResult.RetMsg)
				}
				return nil
			}
		case <-ticker.C:
			s.tx <- b
		case <-timeout:
			return errors.New("timeout")
		}
	}
}

// type Stream struct {
// 	conn *ws.Conn
//
// 	topics map[string]*utils.Broadcaster[*models.StreamData]
// 	ops    map[string]*utils.Future[error]
//
// 	mu         sync.RWMutex
// 	wg         sync.WaitGroup
// 	subBarrier sync.Mutex
//
// 	ctx    context.Context
// 	cancel context.CancelFunc
// }
//
// func NewStream(
// 	ctx context.Context,
// 	url string,
// 	onOpen func(*ws.Conn) error,
// ) *Stream {
// 	ctx, cancel := context.WithCancel(ctx)
//
// 	s := &Stream{
// 		topics: make(map[string]*utils.Broadcaster[*models.StreamData]),
// 		ops:    make(map[string]*utils.Future[error]),
//
// 		ctx:    ctx,
// 		cancel: cancel,
// 	}
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
// 			models.StreamOpSubscribe,
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
// func (s *Stream) Topics() []string {
// 	s.mu.RLock()
// 	defer s.mu.RUnlock()
//
// 	out := make([]string, 0, len(s.topics))
// 	for t := range s.topics {
// 		out = append(out, t)
// 	}
//
// 	return out
// }
//
// func (s *Stream) Subscribe(topics []string, buffer int) (*utils.Subscription[*models.StreamData], error) {
// 	if len(topics) == 0 {
// 		return nil, fmt.Errorf("empty topics")
// 	}
// 	if err := s.ctx.Err(); err != nil {
// 		return nil, err
// 	}
//
// 	s.subBarrier.Lock()
// 	defer s.subBarrier.Unlock()
//
// 	set := make(map[string]struct{}, len(topics))
// 	for _, t := range s.Topics() {
// 		set[t] = struct{}{}
// 	}
//
// 	newTopics := make([]string, 0)
// 	for _, t := range topics {
// 		if _, ok := set[t]; !ok {
// 			newTopics = append(newTopics, t)
// 		}
// 	}
//
// 	if len(newTopics) > 0 {
// 		reqID, b, err := newStreamOpRequest(models.StreamOpSubscribe, newTopics)
// 		if err != nil {
// 			return nil, err
// 		}
//
// 		fut := utils.NewFuture[error](func() {
// 			s.mu.Lock()
// 			delete(s.ops, reqID)
// 			s.mu.Unlock()
// 		})
//
// 		defer fut.Complete(nil)
//
// 		s.mu.Lock()
// 		s.ops[reqID] = fut
// 		s.mu.Unlock()
//
// 		if err := s.conn.Send(b); err != nil {
// 			return nil, err
// 		}
//
// 		select {
// 		case err := <-fut.Await():
// 			if err != nil {
// 				return nil, err
// 			}
// 		case <-time.After(5 * time.Second):
// 			return nil, errors.New("subscribe timeout")
// 		case <-s.ctx.Done():
// 			return nil, errors.New("stream closed")
// 		}
// 	}
//
// 	subs := make([]*utils.Subscription[*models.StreamData], 0)
//
// 	s.mu.Lock()
// 	for _, t := range topics {
// 		if s.topics[t] == nil {
// 			s.topics[t] = utils.NewBroadcaster[*models.StreamData]()
// 		}
// 		sub := s.topics[t].Subscribe(buffer, func(b *utils.Broadcaster[*models.StreamData]) {
// 			if b.IsEmpty() {
// 				s.subBarrier.Lock()
// 				defer s.subBarrier.Unlock()
//
// 				s.mu.Lock()
// 				delete(s.topics, t)
// 				s.mu.Unlock()
//
// 				_, b, err := newStreamOpRequest(models.StreamOpUnsubscribe, []string{t})
// 				if err != nil {
// 					return
// 				}
//
// 				s.conn.Send(b)
// 			}
// 		})
// 		subs = append(subs, sub)
// 	}
// 	s.mu.Unlock()
//
// 	return utils.MergeSubscriptions(buffer, subs...), nil
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
//
// 			var data models.StreamData
// 			if json.Unmarshal(b, &data) == nil {
//
// 				s.mu.RLock()
// 				if bc, ok := s.topics[data.Topic]; ok {
// 					bc.Publish(&data)
// 				}
// 				s.mu.RUnlock()
//
// 				continue
// 			}
// 		}
//
// 		var res models.StreamOpResult
// 		if json.Unmarshal(b, &res) == nil {
//
// 			s.mu.RLock()
// 			fu := s.ops[res.ReqID]
// 			s.mu.RUnlock()
//
// 			if fu != nil {
// 				if res.Success {
// 					fu.Complete(nil)
// 					continue
// 				}
// 				fu.Complete(errors.New(res.RetMsg))
// 			}
// 		}
// 	}
// }
//
// func (s *Stream) Close() {
// 	s.subBarrier.Lock()
// 	s.cancel()
// 	s.wg.Wait()
// 	s.subBarrier.Unlock()
//
// 	for _, b := range s.topics {
// 		b.Clear()
// 	}
// }

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

	"github.com/gorilla/websocket"
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
	onConnected ws.OnConnectedFn,
	privateAuth ws.OnConnectedFn,

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
			func(conn *websocket.Conn, sendCh chan<- []byte, n int) error {

				if privateAuth != nil {
					if err := privateAuth(conn, sendCh, n); err != nil {
						return err
					}
				}

				if n > 0 {
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
				}

				if onConnected != nil {
					if err := onConnected(conn, sendCh, n); err != nil {
						return err
					}
				}

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

	ticker := time.NewTicker(time.Second * 5)
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

	ticker := time.NewTicker(time.Second * 5)
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

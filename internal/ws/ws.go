package ws

import (
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type MessageHandlerFn func(int, []byte) (any, bool)
type OnConnectedFn func(chan<- []byte, int) error
type OnConnectFn func(int)
type OnDisconnectFn func(int, bool)

type Policy struct {
	Dialer *websocket.Dialer

	OnConnected    OnConnectedFn
	MessageHandler MessageHandlerFn
	OnConnect      OnConnectFn
	OnDisconnect   OnDisconnectFn

	ReconnectTimeout time.Duration
	PingInterval     time.Duration
	WriteTimeout     time.Duration
}

type PolicyOption func(*Policy)

func NewPolicy(

	onConnected OnConnectedFn,
	messageHandler MessageHandlerFn,

	opts ...PolicyOption,

) *Policy {

	p := &Policy{
		Dialer: &websocket.Dialer{
			HandshakeTimeout: 7 * time.Second,
		},

		OnConnected:    onConnected,
		MessageHandler: messageHandler,
		OnConnect:      nil,
		OnDisconnect:   nil,

		ReconnectTimeout: 200 * time.Millisecond,
		PingInterval:     15 * time.Second,
		WriteTimeout:     10 * time.Second,
	}

	for _, opt := range opts {
		opt(p)
	}

	return p
}

func WithDialer(d *websocket.Dialer) PolicyOption {
	return func(p *Policy) {
		p.Dialer = d
	}
}

func WithOnDisconnectFn(f OnDisconnectFn) PolicyOption {
	return func(p *Policy) {
		p.OnDisconnect = f
	}
}

func WithOnConnectFn(f OnConnectFn) PolicyOption {
	return func(p *Policy) {
		p.OnConnect = f
	}
}

func WithReconnectTimeout(d time.Duration) PolicyOption {
	return func(p *Policy) {
		p.ReconnectTimeout = d
	}
}

func WithPingInterval(d time.Duration) PolicyOption {
	return func(p *Policy) {
		p.PingInterval = d
	}
}

func WithWriteTimeout[T any](d time.Duration) PolicyOption {
	return func(p *Policy) {
		p.WriteTimeout = d
	}
}

func NewConnV2(

	url string,
	header http.Header,
	tx chan []byte,
	rxBuff int,
	policy *Policy,

) <-chan any {

	rx := make(chan any, rxBuff)

	go func() {
		defer close(rx)

		var wg sync.WaitGroup
		exitF := false
		n := 0

		for {
			wg.Wait()

			if exitF {
				return
			}

			conn, _, err := policy.Dialer.Dial(url, header)
			if err != nil {
				time.Sleep(policy.ReconnectTimeout)
				continue
			}

			n++

			if policy.OnConnect != nil {
				policy.OnConnect(n - 1)
			}

			done := make(chan struct{}, 1)
			wg.Go(func() {
				ticker := time.NewTicker(policy.PingInterval)
				defer ticker.Stop()

			loop:
				for {
					select {

					case b, ok := <-tx:

						// reconnect if data is nil
						if ok && b == nil {
							break loop
						}

						conn.SetWriteDeadline(time.Now().Add(policy.WriteTimeout))
						if !ok {
							conn.WriteMessage(
								websocket.CloseMessage,
								websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""),
							)

							exitF = true
							break loop
						}

						if err := conn.WriteMessage(websocket.TextMessage, b); err != nil {
							break loop
						}

					case <-ticker.C:

						conn.SetWriteDeadline(time.Now().Add(policy.WriteTimeout))
						if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
							break loop
						}

					case <-done:
						break loop

					}
				}

				conn.Close()

				if policy.OnDisconnect != nil {
					policy.OnDisconnect(n-1, exitF)
				}

				if !exitF {
					time.Sleep(policy.ReconnectTimeout)
				}
			})

			if policy.OnConnected != nil {
				err := policy.OnConnected(tx, n-1)
				if err != nil {
					done <- struct{}{}
					continue
				}
			}

			conn.SetReadDeadline(time.Now().Add(policy.PingInterval * 2))
			conn.SetPongHandler(func(string) error {
				conn.SetReadDeadline(time.Now().Add(policy.PingInterval * 2))
				return nil
			})

			conn.SetPingHandler(func(string) error {
				conn.SetWriteDeadline(time.Now().Add(policy.WriteTimeout))
				return conn.WriteMessage(websocket.PongMessage, nil)
			})

			for {
				mt, b, err := conn.ReadMessage()
				if err != nil {
					done <- struct{}{}
					break
				}

				if d, skip := policy.MessageHandler(mt, b); !skip {
					rx <- d
				}
			}
		}
	}()

	return rx
}

// const (
// 	pingInterval = 20 * time.Second
// 	retryDelay   = time.Second
// 	writeTimeout = 5 * time.Second
// )
//
// type Conn struct {
// 	tx chan []byte
// 	rx chan []byte
//
// 	url    string
// 	onOpen func(*Conn) error
//
// 	wg     sync.WaitGroup
// 	ctx    context.Context
// 	cancel context.CancelFunc
// }
//
// func NewConn(ctx context.Context, url string, onOpen func(*Conn) error) *Conn {
// 	ctx, cancel := context.WithCancel(ctx)
//
// 	c := &Conn{
// 		tx: make(chan []byte, 128),
// 		rx: make(chan []byte, 256),
//
// 		url:    url,
// 		onOpen: onOpen,
//
// 		ctx:    ctx,
// 		cancel: cancel,
// 	}
//
// 	c.wg.Go(c.run)
//
// 	return c
// }
//
// func (c *Conn) Send(b []byte) error {
// 	// if err := c.ctx.Err(); err != nil {
// 	// 	return err
// 	// }
// 	select {
// 	case c.tx <- b:
// 		return nil
// 	case <-c.ctx.Done():
// 		return c.ctx.Err()
// 	}
// }
//
// func (c *Conn) Recv() ([]byte, error) {
// 	if b, ok := <-c.rx; ok {
// 		return b, nil
// 	}
// 	return nil, c.ctx.Err()
// }
//
// func (c *Conn) Close() {
// 	c.cancel()
// 	c.wg.Wait()
// }
//
// func (c *Conn) Err() error {
// 	return c.ctx.Err()
// }
//
// func (c *Conn) run() {
// 	for {
// 		var done chan struct{}
// 		var ticker *time.Ticker
//
// 		dialer := &websocket.Dialer{
// 			EnableCompression: false,
// 			HandshakeTimeout:  writeTimeout,
// 		}
//
// 		conn, _, err := dialer.Dial(c.url, nil)
// 		if err != nil {
// 			goto reconnect
// 		}
//
// 		done = make(chan struct{})
// 		go func() {
// 			conn.SetReadDeadline(time.Now().Add(pingInterval * 2))
// 			conn.SetPongHandler(func(string) error {
// 				conn.SetReadDeadline(time.Now().Add(pingInterval * 2))
// 				return nil
// 			})
//
// 			for {
// 				_, b, err := conn.ReadMessage()
// 				if err != nil {
// 					close(done)
// 					return
// 				}
//
// 				select {
// 				case c.rx <- b:
// 				case <-c.ctx.Done():
// 					return
// 				}
// 			}
// 		}()
//
// 		if c.onOpen != nil {
// 			if err := c.onOpen(c); err != nil {
// 				goto reconnect
// 			}
// 		}
//
// 		ticker = time.NewTicker(pingInterval)
// 		for {
// 			select {
// 			case b := <-c.tx:
// 				conn.SetWriteDeadline(time.Now().Add(writeTimeout))
// 				if err := conn.WriteMessage(websocket.TextMessage, b); err != nil {
// 					goto reconnect
// 				}
// 			case <-ticker.C:
// 				conn.SetWriteDeadline(time.Now().Add(writeTimeout))
// 				if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
// 					goto reconnect
// 				}
// 			case <-done:
// 				goto reconnect
// 			case <-c.ctx.Done():
// 				_ = conn.WriteMessage(
// 					websocket.CloseMessage,
// 					websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""),
// 				)
// 				goto reconnect
// 			}
// 		}
//
// 	reconnect:
// 		if conn != nil {
// 			conn.Close()
// 		}
// 		if ticker != nil {
// 			ticker.Stop()
// 		}
// 		if c.ctx.Err() != nil {
// 			return
// 		}
// 		time.Sleep(retryDelay)
//
// 	}
// }

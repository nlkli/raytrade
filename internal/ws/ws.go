package ws

import (
	"context"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	pingInterval = 20 * time.Second
	retryDelay   = time.Second
	writeTimeout = 5 * time.Second
)

type Conn struct {
	tx chan []byte
	rx chan []byte

	url    string
	onOpen func(*Conn) error

	wg     sync.WaitGroup
	ctx    context.Context
	cancel context.CancelFunc
}

func NewConn(ctx context.Context, url string, onOpen func(*Conn) error) *Conn {
	ctx, cancel := context.WithCancel(ctx)

	c := &Conn{
		tx: make(chan []byte, 128),
		rx: make(chan []byte, 256),

		url:    url,
		onOpen: onOpen,

		ctx:    ctx,
		cancel: cancel,
	}

	c.wg.Go(c.run)

	return c
}

func (c *Conn) Send(b []byte) error {
	if err := c.ctx.Err(); err != nil {
		return err
	}
	select {
	case c.tx <- b:
		return nil
	case <-c.ctx.Done():
		return c.ctx.Err()
	}
}

func (c *Conn) Recv() ([]byte, error) {
	select {
	case b := <-c.rx:
		return b, nil
	case <-c.ctx.Done():
		return nil, c.ctx.Err()
	}
}

func (c *Conn) Close() {
	c.cancel()
	c.wg.Wait()
}

func (c *Conn) Err() error {
	return c.ctx.Err()
}

func (c *Conn) run() {
	for {
		var done chan struct{}
		var ticker *time.Ticker
		conn, _, err := websocket.DefaultDialer.Dial(c.url, nil)
		if err != nil {
			goto reconnect
		}

		done = make(chan struct{})
		go func() {
			conn.SetReadDeadline(time.Now().Add(pingInterval * 2))
			conn.SetPongHandler(func(string) error {
				conn.SetReadDeadline(time.Now().Add(pingInterval * 2))
				return nil
			})

			for {
				_, b, err := conn.ReadMessage()
				if err != nil {
					close(done)
					return
				}

				select {
				case c.rx <- b:
				case <-c.ctx.Done():
					return
				}
			}
		}()

		if c.onOpen != nil {
			if err := c.onOpen(c); err != nil {
				goto reconnect
			}
		}

		ticker = time.NewTicker(pingInterval)
		for {
			select {
			case b := <-c.tx:
				conn.SetWriteDeadline(time.Now().Add(writeTimeout))
				if err := conn.WriteMessage(websocket.TextMessage, b); err != nil {
					goto reconnect
				}
			case <-ticker.C:
				conn.SetWriteDeadline(time.Now().Add(writeTimeout))
				if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
					goto reconnect
				}
			case <-done:
				goto reconnect
			case <-c.ctx.Done():
				_ = conn.WriteMessage(
					websocket.CloseMessage,
					websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""),
				)
				goto reconnect
			}
		}

	reconnect:
		if conn != nil {
			conn.Close()
		}
		if ticker != nil {
			ticker.Stop()
		}
		if c.ctx.Err() != nil {
			return
		}
		time.Sleep(retryDelay)

	}
}

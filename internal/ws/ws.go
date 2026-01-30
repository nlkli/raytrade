package ws

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type Conn struct {
	tx chan []byte
	rx chan []byte

	onOpen func(*Conn) error

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

func NewConn(ctx context.Context, url string, onOpen func(*Conn) error) *Conn {
	ctx, cancel := context.WithCancel(ctx)

	c := &Conn{
		tx: make(chan []byte, 128),
		rx: make(chan []byte, 256),

		onOpen: onOpen,

		ctx:    ctx,
		cancel: cancel,
	}

	c.wg.Add(1)
	go c.run(url)

	return c
}

func (c *Conn) Send(msg []byte) error {
	if len(msg) == 0 {
		return errors.New("empty message")
	}

	select {
	case c.tx <- msg:
		return nil
	case <-c.ctx.Done():
		return c.ctx.Err()
	}
}

func (c *Conn) Recv() ([]byte, error) {
	select {
	case msg := <-c.rx:
		return msg, nil
	case <-c.ctx.Done():
		return nil, c.ctx.Err()
	}
}

func (c *Conn) Close() {
	c.cancel()
	c.wg.Wait()
}

func (c *Conn) run(url string) {
	defer c.wg.Done()

	const (
		pingInterval = 20 * time.Second
		retryDelay   = time.Second
		writeTimeout = 5 * time.Second
	)

	for {
		select {
		case <-c.ctx.Done():
			return
		default:
		}

		conn, _, err := websocket.DefaultDialer.Dial(url, nil)
		if err != nil {
			time.Sleep(retryDelay)
			continue
		}

		if c.onOpen != nil {
			if err := c.onOpen(c); err != nil {
				conn.Close()
				time.Sleep(retryDelay)
				continue
			}
		}

		readDone := make(chan struct{})

		go func() {
			defer close(readDone)

			conn.SetReadDeadline(time.Now().Add(pingInterval * 2))
			conn.SetPongHandler(func(string) error {
				conn.SetReadDeadline(time.Now().Add(pingInterval * 2))
				return nil
			})

			for {
				_, msg, err := conn.ReadMessage()
				if err != nil {
					return
				}

				select {
				case c.rx <- msg:
				case <-c.ctx.Done():
					return
				}
			}
		}()

		ticker := time.NewTicker(pingInterval)

		for {
			select {
			case msg := <-c.tx:
				conn.SetWriteDeadline(time.Now().Add(writeTimeout))
				if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
					goto reconnect
				}

			case <-ticker.C:
				conn.SetWriteDeadline(time.Now().Add(writeTimeout))
				if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
					goto reconnect
				}

			case <-readDone:
				goto reconnect

			case <-c.ctx.Done():
				goto shutdown
			}
		}

	reconnect:
		ticker.Stop()
		conn.Close()
		time.Sleep(retryDelay)
		continue

	shutdown:
		ticker.Stop()
		_ = conn.WriteMessage(
			websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""),
		)
		conn.Close()
		return
	}
}

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

	// state   map[string][]byte
	// stateMu sync.Mutex

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	onOpen func(*Conn) error
}

func NewConn(ctx context.Context, url string, onOpen func(*Conn) error) *Conn {
	ctx, cancel := context.WithCancel(ctx)

	c := &Conn{
		tx: make(chan []byte, 128),
		rx: make(chan []byte, 256),
		// state:  make(map[string][]byte),
		ctx:    ctx,
		cancel: cancel,
		onOpen: onOpen,
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

// func (c *Conn) SendPersistent(key string, msg []byte) error {
// 	if len(msg) == 0 {
// 		return errors.New("empty message")
// 	}
// 	c.stateMu.Lock()
// 	c.state[key] = msg
// 	c.stateMu.Unlock()
//
// 	return c.Send(msg)
// }
//
// func (c *Conn) ClearPersistent(key string) {
// 	c.stateMu.Lock()
// 	delete(c.state, key)
// 	c.stateMu.Unlock()
// }

func (c *Conn) Recv() ([]byte, error) {
	select {
	case msg, ok := <-c.rx:
		if !ok {
			return nil, errors.New("connection closed")
		}
		return msg, nil
	case <-c.ctx.Done():
		return nil, c.ctx.Err()
	}
}

func (c *Conn) Close() {
	c.cancel()
	c.wg.Wait()
	close(c.tx)
	close(c.rx)
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

		// if err := c.replayState(conn); err != nil {
		// 	conn.Close()
		// 	time.Sleep(retryDelay)
		// 	continue
		// }

		done := make(chan struct{})

		go func() {
			defer close(done)

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

	loop:
		for {
			select {
			case msg := <-c.tx:
				conn.SetWriteDeadline(time.Now().Add(writeTimeout))
				if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
					break loop
				}

			case <-ticker.C:
				conn.SetWriteDeadline(time.Now().Add(writeTimeout))
				if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
					break loop
				}

			case <-done:
				break loop

			case <-c.ctx.Done():
				break loop
			}
		}

		ticker.Stop()
		_ = conn.WriteMessage(
			websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""),
		)
		conn.Close()
	}
}

// func (c *Conn) replayState(conn *websocket.Conn) error {
// 	c.stateMu.Lock()
// 	defer c.stateMu.Unlock()
// 
// 	for _, msg := range c.state {
// 		conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
// 		if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
// 			return err
// 		}
// 	}
// 	return nil
// }

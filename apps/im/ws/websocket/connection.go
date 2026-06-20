package websocket

import (
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type Conn struct {
	idlemu sync.Mutex // 空闲操作的锁
	*websocket.Conn
	s    *Server
	idle time.Time

	maxConnections time.Duration

	done chan struct{}
}

func NewConn(s *Server, w http.ResponseWriter, r *http.Request) *Conn {
	c, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		s.Errorf("upgrade err %v", err)
		return nil
	}

	conn := &Conn{
		Conn:           c,
		s:              s,
		idle:           time.Now(),
		maxConnections: defaultMaxConnectionIdle,
		done:           make(chan struct{}),
	}

	return conn
}

func (c *Conn) ReadMessage() (messageType int, p []byte, err error) {
	messageType, p, err = c.Conn.ReadMessage()

	c.idlemu.Lock()
	defer c.idlemu.Unlock()

	c.idle = time.Time{}
	return
}

func (c *Conn) WriteMessage(messageType int, data []byte) error {
	// 在这里第三方库协程并不安全
	c.idlemu.Lock()
	defer c.idlemu.Unlock()

	err := c.Conn.WriteMessage(messageType, data)
	c.idle = time.Now()
	return err
}

func (c *Conn) Close() error {
	select {
	case <-c.done:
	default:
		close(c.done)
	}

	return c.Conn.Close()
}

func (c *Conn) keepalive() {
	idleTimer := time.NewTimer(c.maxConnections)
	defer func() {
		idleTimer.Stop()
	}()

	for {
		select {
		case <-idleTimer.C:
			c.idlemu.Lock()
			idle := c.idle
			if idle.IsZero() { // The connection is non-idle.
				c.idlemu.Unlock()
				idleTimer.Reset(c.maxConnections)
				continue
			}
			val := c.maxConnections - time.Since(idle)
			c.idlemu.Unlock()
			if val <= 0 {
				// The connection has been idle for a duration of keepalive.MaxConnectionIdle or more.
				// Gracefully close the connection.
				c.s.Close(c)
				return
			}
			idleTimer.Reset(val)
		}
	}
}

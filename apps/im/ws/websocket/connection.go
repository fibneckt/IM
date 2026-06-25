package websocket

import (
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type Conn struct {
	// 保护 idle 字段的并发读写
	idleMu sync.Mutex
	// 连接对应的用户id
	Uid string
	// 嵌入
	*websocket.Conn
	// 反向引用Server，用于 keepalive 超时后调用 s.Close(c) 清理
	s *Server
	// 最后一次读/写的时间戳
	idle time.Time
	// 最大空闲时长，从 serverOption 传入，默认10s
	maxConnectionIdle time.Duration

	// 读消息队列
	readMessage []*Message
	// 序列化 key - 消息id Message - 具体消息
	readMessageSeq map[string]*Message
	// 并发保护消息读写
	messageMu sync.Mutex
	// 用于ACK确认之后将消息发送给任务处理
	message chan *Message

	// 关闭信号
	done chan struct{}
}

func NewConn(s *Server, w http.ResponseWriter, r *http.Request) *Conn {
	c, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		s.Errorf("upgrade err %v", err)
		return nil
	}

	conn := &Conn{
		Conn:              c,
		s:                 s,
		idle:              time.Now(),
		maxConnectionIdle: s.opt.maxConnectionIdle,
		done:              make(chan struct{}),
		readMessage:       make([]*Message, 0, 2),
		readMessageSeq:    make(map[string]*Message, 2),
		// 容量为1 确保收和发的顺序
		message: make(chan *Message, 1),
	}

	go conn.keepalive()
	return conn
}

func (c *Conn) ReadMessage() (messageType int, p []byte, err error) {
	messageType, p, err = c.Conn.ReadMessage()

	c.idleMu.Lock()
	defer c.idleMu.Unlock()
	c.idle = time.Now()
	return
}

func (c *Conn) WriteMessage(messageType int, data []byte) error {
	c.idleMu.Lock()
	defer c.idleMu.Unlock()

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
	ticker := time.NewTicker(c.maxConnectionIdle / 2)
	defer ticker.Stop()

	for {
		select {
		case <-c.done:
			return
		case <-ticker.C:
			c.idleMu.Lock()
			idle := c.idle
			c.idleMu.Unlock()

			if time.Since(idle) >= c.maxConnectionIdle {
				c.s.Close(c)
				return
			}
		}
	}
}

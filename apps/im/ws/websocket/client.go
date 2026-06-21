package websocket

import (
	"encoding/json"
	"net/url"

	"github.com/gorilla/websocket"
)

// websocket 客户端

type Client interface {
	Close() error

	Send(v any) error
	Read(v any) error
}

type client struct {
	// 对应 websocket 连接
	*websocket.Conn
	// 连接服务的地址
	host string

	opt dailOption
}

func (c *client) Send(v any) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}

	err = c.WriteMessage(websocket.TextMessage, data)
	if err != nil {
		return err
	}
	// todo: 再添加一个重连发送
	conn, err := c.dail()
	if err != nil {
		return err
	}
	c.Conn = conn
	return c.WriteMessage(websocket.TextMessage, data)
}

func (c *client) Read(v any) error {
	_, msg, err := c.Conn.ReadMessage()
	if err != nil {
		return err
	}
	// 仅仅做一个序列化的转换
	return json.Unmarshal(msg, v)
}

func NewClient(host string, opts ...DailOptions) *client {
	opt := newDailOptions(opts...)
	c := &client{
		Conn: nil,
		host: host,
		opt:  opt,
	}

	conn, err := c.dail()
	if err != nil {
		panic(err)
	}
	c.Conn = conn
	return c
}

// 用于建立客户端与服务端的连接
func (c *client) dail() (*websocket.Conn, error) {
	u := url.URL{Scheme: "ws", Host: c.host, Path: c.opt.pattern}
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	return conn, err
}

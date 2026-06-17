package websocket

import "github.com/gorilla/websocket"

type Route struct {
	Method  string
	Handler HandlerFunc
}

// 定义路由
type HandlerFunc func(srv *Server, conn *websocket.Conn, msg *Message)

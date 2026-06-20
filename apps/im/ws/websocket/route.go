package websocket

type Route struct {
	Method  string
	Handler HandlerFunc
}

// 定义路由
type HandlerFunc func(srv *Server, conn *Conn, msg *Message)

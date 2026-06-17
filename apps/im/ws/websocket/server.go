package websocket

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/zeromicro/go-zero/core/logx"
)

type Server struct {
	routers map[string]HandlerFunc // 路由

	addr     string
	upgrader websocket.Upgrader
	logx.Logger
}

func NewServer(addr string) *Server {
	return &Server{
		routers:  make(map[string]HandlerFunc),
		addr:     addr,
		upgrader: websocket.Upgrader{},
		Logger:   logx.WithContext(context.Background()),
	}
}

// 处理请求方法
func (s *Server) ServerWs(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if r := recover(); r != nil {
			s.Errorf("server handler ws recover err %v", r)
		}
	}()

	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		s.Errorf("Upgrader err %v", err)
		return
	}

	// 根据连接对象获取请求信息
	// method
	go s.handlerConn(conn)
}

func (s *Server) handlerConn(conn *websocket.Conn) {
	for {
		// 获取请求消息
		_, msg, err := conn.ReadMessage()
		if err != nil {
			s.Errorf("websocket conn read message err %v", err)
			// todo : turn off connect
			return
		}
		var message Message
		if err = json.Unmarshal(msg, &message); err != nil {
			s.Errorf("json unmarshal err %v, msg %v", err, string(msg))
			// todo : turn off connect
			return
		}

		// 根据请求的 method 分发路由并执行
		if handler, ok := s.routers[message.Method]; ok {
			handler(s, conn, &message)
		} else {
			conn.WriteMessage(websocket.CloseMessage, []byte(fmt.Sprintf("不存在的执行方法 %v 请检查", message.Method)))
		}
	}
}

func (s *Server) AddRoutes(rs []Route) {
	for _, r := range rs {
		s.routers[r.Method] = r.Handler
	}
}

// 启动服务
func (s *Server) Start() {
	http.HandleFunc("/ws", s.ServerWs)
	s.Info(http.ListenAndServe(s.addr, nil))
}

// 停止服务
func (s *Server) Stop() {
	fmt.Println("server stop")
}

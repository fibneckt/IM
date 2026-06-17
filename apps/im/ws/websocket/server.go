package websocket

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/zeromicro/go-zero/core/logx"
)

type Server struct {
	routers map[string]HandlerFunc // 路由
	addr    string

	sync.RWMutex                              // 加锁防止下面两个 map 并发出现的错误
	connToUser     map[*websocket.Conn]string // 连接对应的 user
	userToConn     map[string]*websocket.Conn // user 对应的连接
	authentication Authentication

	upgrader websocket.Upgrader
	logx.Logger
}

func NewServer(addr string) *Server {
	return &Server{
		routers:        make(map[string]HandlerFunc),
		addr:           addr,
		authentication: new(authentication),
		userToConn:     make(map[string]*websocket.Conn),
		connToUser:     make(map[*websocket.Conn]string),

		upgrader: websocket.Upgrader{},
		Logger:   logx.WithContext(context.Background()),
	}
}

// 获取请求id
func (s *Server) addConn(conn *websocket.Conn, req *http.Request) {
	uid := s.authentication.UserId(req)

	// 配置锁
	s.RWMutex.Lock()
	defer s.RWMutex.Unlock()

	s.connToUser[conn] = uid
	s.userToConn[uid] = conn
}

func (s *Server) GetConns(uids ...string) []*websocket.Conn {
	if len(uids) == 0 {
		return nil
	}

	s.RWMutex.RLock()
	defer s.RWMutex.RUnlock()

	res := make([]*websocket.Conn, 0, len(uids))
	for _, uid := range uids {
		res = append(res, s.userToConn[uid])
	}
	return res
}

func (s *Server) GetUsers(conns ...*websocket.Conn) []string {

	s.RWMutex.RLock()
	defer s.RWMutex.RUnlock()

	var res []string
	if len(conns) == 0 {
		// 获取全部
		res = make([]string, 0, len(s.connToUser))
		for _, uid := range s.connToUser {
			res = append(res, uid)
		}
	} else {
		// 获取部分
		res = make([]string, 0, len(conns))
		for _, conn := range conns {
			res = append(res, s.connToUser[conn])
		}
	}

	return res
}

func (s *Server) GetConn(uid string) *websocket.Conn {
	s.RWMutex.RLock()
	defer s.RWMutex.RUnlock()
	return s.userToConn[uid]
}

func (s *Server) GetUser(uid string) *websocket.Conn {
	s.RWMutex.RLock()
	defer s.RWMutex.RUnlock()
	return s.userToConn[uid]
}

func (s *Server) Close(conn *websocket.Conn) {
	s.RWMutex.Lock()
	defer s.RWMutex.Unlock()
	uid := s.connToUser[conn]
	if uid != "" {
		// 已经被关闭
		return
	}
	delete(s.connToUser, conn)
	delete(s.userToConn, uid)

	conn.Close()
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
			s.Close(conn)
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

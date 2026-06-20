package websocket

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/zeromicro/go-zero/core/logx"
)

type Server struct {
	routers map[string]HandlerFunc // 路由
	addr    string

	patten string

	sync.RWMutex                    // 加锁防止下面两个 map 并发出现的错误
	connToUser     map[*Conn]string // 连接对应的 user
	userToConn     map[string]*Conn // user 对应的连接
	authentication Authentication

	upgrader websocket.Upgrader
	logx.Logger
	opt *serverOption
}

func NewServer(addr string, opts ...ServerOptions) *Server {
	opt := newServerOptions(opts...)

	return &Server{
		routers: make(map[string]HandlerFunc),
		addr:    addr,
		opt:     &opt,
		// 头疼测试时用了 new(authentication) 写死了返回 true，用时间戳当 userId 的默认实现
		authentication: opt.Authentication,
		userToConn:     make(map[string]*Conn),
		connToUser:     make(map[*Conn]string),

		patten: opt.patten,

		upgrader: websocket.Upgrader{},
		Logger:   logx.WithContext(context.Background()),
	}
}

// 获取请求id
func (s *Server) addConn(conn *Conn, req *http.Request) {
	uid := s.authentication.UserId(req)

	// 配置锁
	s.RWMutex.Lock()
	defer s.RWMutex.Unlock()

	// 某些用户可能重复登录，这里进行验证
	if c := s.userToConn[uid]; c != nil {
		// 关闭之前的连接
		c.Close()
	}

	s.connToUser[conn] = uid
	s.userToConn[uid] = conn
}

func (s *Server) GetConns(uids ...string) []*Conn {
	if len(uids) == 0 {
		return nil
	}

	s.RWMutex.RLock()
	defer s.RWMutex.RUnlock()

	res := make([]*Conn, 0, len(uids))
	for _, uid := range uids {
		res = append(res, s.userToConn[uid])
	}
	return res
}

func (s *Server) GetUsers(conns ...*Conn) []string {

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

func (s *Server) GetConn(uid string) *Conn {
	s.RWMutex.RLock()
	defer s.RWMutex.RUnlock()
	return s.userToConn[uid]
}

func (s *Server) GetUser(uid string) *Conn {
	s.RWMutex.RLock()
	defer s.RWMutex.RUnlock()
	return s.userToConn[uid]
}

func (s *Server) Close(conn *Conn) {
	s.RWMutex.Lock()
	defer s.RWMutex.Unlock()
	uid := s.connToUser[conn]
	if uid == "" {
		// 连接未注册，直接关闭
		conn.Conn.Close()
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

	conn := NewConn(s, w, r)
	if conn == nil {
		return
	}

	if !s.authentication.Auth(w, r) {
		s.Send(&Message{FrameType: FrameData, Data: fmt.Sprintf("不具备访问权限")}, conn)
		conn.Close()
		return
	}

	// 记录连接
	s.addConn(conn, r)

	// 根据连接对象获取请求信息
	// method
	go s.handlerConn(conn)
}

func (s *Server) handlerConn(conn *Conn) {
	for {
		// 设置读超时：10s 内没收到消息（包括 FramePing），主动断开
		if err := conn.Conn.SetReadDeadline(time.Now().Add(s.opt.maxConnectionIdle)); err != nil {
			s.Errorf("set read deadline err %v", err)
			s.Close(conn)
			return
		}

		_, msg, err := conn.ReadMessage()
		if err != nil {
			s.Errorf("websocket conn read message err %v", err)
			s.Close(conn)
			return
		}

		// 解析消息
		var message Message
		if err = json.Unmarshal(msg, &message); err != nil {
			s.Errorf("json unmarshal err %v, msg %v", err, string(msg))
			s.Close(conn)
			return
		}

		// 根据消息类型进行处理
		switch message.FrameType {
		case FramePing:
			// 心跳响应：收到 ping 回复 pong，同时读超时已被重置
			s.Send(&Message{FrameType: FramePing}, conn)
		case FrameData:
			// 根据请求的 method 分发路由并执行
			if handler, ok := s.routers[message.Method]; ok {
				handler(s, conn, &message)
			} else {
				s.Send(&Message{FrameType: FrameData, Data: fmt.Sprintf("不存在执行的方法 %v 请检查", message.Method)}, conn)
			}
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
	http.HandleFunc(s.patten, s.ServerWs)
	s.Info(http.ListenAndServe(s.addr, nil))
}

// 停止服务
func (s *Server) Stop() {
	fmt.Println("server stop")
}

func (s *Server) SendByUserId(msg interface{}, sendIds ...string) error {
	if len(sendIds) == 0 {
		return nil
	}

	return s.Send(msg, s.GetConns(sendIds...)...)
}

func (s *Server) Send(msg interface{}, conns ...*Conn) error {
	if len(conns) == 0 {
		return nil
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	for _, conn := range conns {
		if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
			return err
		}
	}
	return nil
}

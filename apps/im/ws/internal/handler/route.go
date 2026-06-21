package handler

import (
	"IM/apps/im/ws/internal/handler/conversation"
	"IM/apps/im/ws/internal/handler/user"
	"IM/apps/im/ws/internal/svc"
	"IM/apps/im/ws/websocket"
)

func RegisterHandlers(srv *websocket.Server, svc *svc.ServiceContext) {
	srv.AddRoutes([]websocket.Route{
		{
			Method:  "user.online",
			Handler: user.OnLine(svc),
		},
		{
			Method:  "conversation.chat",
			Handler: conversation.Chat(svc),
		},
	})
}

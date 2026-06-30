package conversation

import (
	"IM/apps/im/ws/internal/svc"
	"IM/apps/im/ws/websocket"
	"IM/apps/im/ws/ws"
	"IM/apps/task/mq/mq"
	"IM/pkg/constants"
	"time"

	"github.com/mitchellh/mapstructure"
)

func Chat(svc *svc.ServiceContext) websocket.HandlerFunc {
	return func(srv *websocket.Server, conn *websocket.Conn, msg *websocket.Message) {
		// 私聊
		var data ws.Chat
		if err := mapstructure.Decode(msg.Data, &data); err != nil {
			srv.Send(websocket.NewErrMessage(err), conn)
			return
		}

		switch data.ChatType {
		case constants.SingleChatType:

			err := svc.MsgChatTransferClient.Push(&mq.MsgChatTransfer{
				ConversationId: data.ConversationId,
				ChatType:       data.ChatType,
				SendId:         conn.Uid,
				RecvId:         data.ReceiverId,
				SendTime:       time.Now().UnixNano(),
				MType:          data.Msg.MType,
				Content:        data.Msg.Content,
			})
			if err != nil {
				srv.Send(websocket.NewErrMessage(err), conn)
				return
			}
			// 私聊
			//err := logic.NewConversation(context.Background(), srv, svc).SingleChat(&data, conn.Uid)
			//if err != nil {
			//	srv.Send(websocket.NewErrMessage(err), conn)
			//	return
			//}
			//
			//srv.SendByUserId(websocket.NewMessage(conn.Uid, ws.Chat{
			//	ConversationId: data.ConversationId,
			//	ChatType:       data.ChatType,
			//	SenderId:       conn.Uid,
			//	ReceiverId:     data.ReceiverId,
			//	SendTime:       time.Now().UnixMilli(),
			//	Msg:            ws.Msg{},
			//}), data.ReceiverId)

		}
	}
}

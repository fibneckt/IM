package msgTransfer

import (
	"IM/apps/im/immodels"
	"IM/apps/im/ws/websocket"
	"IM/apps/social/rpc/socialclient"
	"IM/apps/task/mq/internal/svc"
	"IM/apps/task/mq/mq"
	"IM/pkg/constants"
	"context"
	"encoding/json"
	"fmt"

	"github.com/zeromicro/go-zero/core/logx"
)

type MsgChatTransfer struct {
	logx.Logger
	svc *svc.ServiceContext // 上下文
}

func (m *MsgChatTransfer) Consume(key, value string) error {
	fmt.Println("key:", key, "value:", value)

	var (
		data mq.MsgChatTransfer
		ctx  = context.Background()
	)

	if err := json.Unmarshal([]byte(value), &data); err != nil {
		return err
	}

	// 记录数据
	if err := m.addChatLog(ctx, &data); err != nil {
		return err
	}

	// 添加对群聊的支持
	switch data.ChatType {
	case constants.GroupChatType:
		return m.group(ctx, &data)
	case constants.SingleChatType:
		return m.single(&data)
	}

	// 推送发送
	return m.svc.WsClient.Send(websocket.Message{
		FrameType: websocket.FrameNoAck,
		Method:    "push",
		FormId:    constants.SYSTEM_ROOT_UID,
		Data:      data,
	})
}

// 消费者
func NewMsgChatTransfer(svc *svc.ServiceContext) *MsgChatTransfer {
	return &MsgChatTransfer{
		Logger: logx.WithContext(context.Background()),
		svc:    svc,
	}
}

func (m *MsgChatTransfer) addChatLog(ctx context.Context, data *mq.MsgChatTransfer) error {
	// 记录消息
	chatLog := immodels.ChatLog{
		ConversationId: data.ConversationId,
		SendId:         data.SendId,
		RecvId:         data.RecvId,
		ChatType:       data.ChatType,
		MsgFrom:        0,
		MsgType:        data.MType,
		MsgContent:     data.Content,
		SendTime:       data.SendTime,
	}
	err := m.svc.ChatLogModel.Insert(ctx, &chatLog)
	if err != nil {
		return err
	}
	return m.svc.ConversationModel.UpdateMsg(ctx, &chatLog)
}

func (m *MsgChatTransfer) single(data *mq.MsgChatTransfer) error {
	return m.svc.WsClient.Send(websocket.Message{
		FrameType: websocket.FrameData,
		Method:    "push",
		FormId:    constants.SYSTEM_ROOT_UID,
		Data:      data,
	})
}

func (m *MsgChatTransfer) group(ctx context.Context, data *mq.MsgChatTransfer) error {
	// 需要查询群用户
	// 使用rpc定义的查询用户方法
	users, err := m.svc.Social.GroupUsers(ctx, &socialclient.GroupUsersReq{
		GroupId: data.RecvId,
	})
	if err != nil {
		return err
	}
	data.RecvIds = make([]string, 0, len(users.List))

	for _, members := range users.List {
		if members.UserId == data.SendId {
			continue
		}
		data.RecvIds = append(data.RecvIds, members.UserId)
	}

	return m.svc.WsClient.Send(websocket.Message{
		FrameType: websocket.FrameData,
		Method:    "push",
		FormId:    constants.SYSTEM_ROOT_UID,
		Data:      data,
	})

}

package msgTransfer

import (
	"IM/apps/im/ws/ws"
	"IM/apps/task/mq/internal/svc"
	"IM/apps/task/mq/mq"
	"IM/pkg/bitmap"
	"IM/pkg/constants"
	"context"
	"encoding/base64"
	"encoding/json"

	"github.com/zeromicro/go-queue/kq"
)

type MsgReadTransfer struct {
	*baseMsgTransfer
}

func NewMsgReadTransfer(svc *svc.ServiceContext) kq.ConsumeHandler {
	return &MsgReadTransfer{
		NewBaseMsgTransfer(svc),
	}
}

func (m *MsgReadTransfer) Consume(key, value string) error {
	m.Info("MsgReadTransfer", value)
	var (
		data mq.MsgMarkRead
		ctx  = context.Background()
	)

	if err := json.Unmarshal([]byte(value), &data); err != nil {
		return err
	}

	// 更新
	readRecords, err := m.UpdateChatLogRead(ctx, &data)
	if err != nil {

	}
	// map[string]string

	return m.Transfer(ctx, &ws.Push{
		ConversationId: data.ConversationId,
		ChatType:       data.ChatType,
		SendId:         data.SendId,
		RecvId:         data.RecvId,
		ReadRecords:    readRecords,
		//SendTime:       data.SendTime,
		//MType:          data.MType,
		//Content:        data.Content,
		//RecvIds:        data.RecvIds,
	})
}

func (m *MsgReadTransfer) UpdateChatLogRead(ctx context.Context, data *mq.MsgMarkRead) (map[string]string, error) {
	res := make(map[string]string)
	chatLogs, err := m.svcCtx.ChatLogModel.ListByMsgIds(ctx, data.MsgIds)
	if err != nil {
		return nil, err
	}

	// 处理已读
	for _, chatLog := range chatLogs {
		switch chatLog.ChatType {
		case constants.SingleChatType:
			chatLog.ReadRecords = []byte{1}
		case constants.GroupChatType:
			readRecords := bitmap.Load(chatLog.ReadRecords)
			readRecords.Set(data.SendId)
			chatLog.ReadRecords = readRecords.Export()
		}
		res[chatLog.ID.Hex()] = base64.StdEncoding.EncodeToString(chatLog.ReadRecords)

		err = m.svcCtx.ChatLogModel.UpdateMakeRead(ctx, chatLog.ID, chatLog.ReadRecords)

		if err != nil {
			return nil, err
		}
	}

	return res, nil
}

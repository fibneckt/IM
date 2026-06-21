package msgTransfer

import (
	"IM/apps/task/mq/internal/svc"
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/logx"
)

type msgChatTransfer struct {
	logx.Logger
	svc *svc.ServiceContext // 上下文
}

func (m *msgChatTransfer) Consume(ctx context.Context, key, value string) error {
	fmt.Println("key:", key, "value:", value)
	return nil
}

// 消费者
func NewMsgChatTransfer(svc *svc.ServiceContext) *msgChatTransfer {
	return &msgChatTransfer{
		Logger: logx.WithContext(context.Background()),
		svc:    svc,
	}
}

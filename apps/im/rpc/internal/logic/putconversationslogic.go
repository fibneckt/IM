package logic

import (
	"IM/apps/im/immodels"
	"IM/pkg/constants"
	"IM/pkg/xerr"
	"context"

	"IM/apps/im/rpc/im"
	"IM/apps/im/rpc/internal/svc"

	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
)

type PutConversationsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewPutConversationsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PutConversationsLogic {
	return &PutConversationsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 更新会话
func (l *PutConversationsLogic) PutConversations(in *im.PutConversationReq) (*im.PutConversationResp, error) {
	conversations, err := l.svcCtx.ConversationsModel.FindByUserId(l.ctx, in.UserId)
	if err != nil {
		return nil, errors.Wrapf(xerr.NewDBErr(), "find conversaion by userId err %v req %v", err, in.UserId)
	}

	if conversations.ConversationList == nil {
		conversations.ConversationList = make(map[string]*immodels.Conversation)
	}

	for k, i := range in.ConversationList {
		var oldTotal int
		if conversations.ConversationList[k] != nil {
			oldTotal = conversations.ConversationList[k].Total
		}

		conversations.ConversationList[k] = &immodels.Conversation{
			ConversationId: i.ConversationId,
			ChatType:       constants.ChatType(i.ChatType),
			IsShow:         i.IsShow,
			// 更新最新的已读总记录
			Total: int(i.Read) + oldTotal,
			Seq:   i.Seq,
		}
	}

	err = l.svcCtx.ConversationsModel.Update(l.ctx, conversations)
	if err != nil {
		return nil, errors.Wrapf(xerr.NewDBErr(), "update conversations err %v req %v", err, in.UserId)
	}

	return &im.PutConversationResp{}, nil
}

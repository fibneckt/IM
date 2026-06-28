package logic

import (
	"IM/apps/im/immodels"
	"IM/pkg/xerr"
	"context"

	"IM/apps/im/rpc/im"
	"IM/apps/im/rpc/internal/svc"

	"github.com/jinzhu/copier"
	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/errorx"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetConversationsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetConversationsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetConversationsLogic {
	return &GetConversationsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取会话
func (l *GetConversationsLogic) GetConversations(in *im.GetConversationReq) (*im.GetConversationResp, error) {
	data, err := l.svcCtx.ConversationsModel.FindByUserId(l.ctx, in.UserId)
	if err != nil {
		if err == immodels.ErrNotFound {
			return &im.GetConversationResp{}, nil
		}
		return nil, errors.Wrapf(xerr.NewDBErr(), "find conversaton by userId err %v req %v", err, in.UserId)
	}

	var res im.GetConversationResp
	copier.Copy(&res, &data)

	ids := make([]string, 0, len(data.ConversationList))
	for _, conversation := range data.ConversationList {
		ids = append(ids, conversation.ConversationId)
	}
	// 统计会话的消息情况
	list, err := l.svcCtx.ConversationModel.ListByConversationIds(l.ctx, ids)
	if err != nil {
		return nil, errorx.Wrapf(xerr.NewDBErr(), "list conversation by ids err %v, req %v", err, ids)
	}

	for _, conversation := range list {
		if _, ok := res.ConversationList[conversation.ConversationId]; !ok {
			continue
		}

		total := res.ConversationList[conversation.ConversationId].Total

		if total < int32(conversation.Total) {
			// 有新的消息
			res.ConversationList[conversation.ConversationId].Total = int32(conversation.Total)
			// 待读消息量
			res.ConversationList[conversation.ConversationId].ToRead = int32(conversation.Total) - total
			// 有新消息一定显示
			res.ConversationList[conversation.ConversationId].IsShow = true
		}
	}

	return &res, nil
}

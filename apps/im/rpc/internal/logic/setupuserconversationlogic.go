package logic

import (
	"IM/apps/im/immodels"
	"IM/apps/im/rpc/im"
	"IM/apps/im/rpc/internal/svc"
	"IM/pkg/constants"
	"IM/pkg/wuid"
	"IM/pkg/xerr"
	"context"

	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type SetUpUserConversationLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSetUpUserConversationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SetUpUserConversationLogic {
	return &SetUpUserConversationLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 建立会话：群聊，私聊
func (l *SetUpUserConversationLogic) SetUpUserConversation(in *im.SetUpUserConversationReq) (*im.SetUpUserConversationResp, error) {
	var res im.SetUpUserConversationResp
	switch constants.ChatType(in.ChatType) {
	case constants.SingleChatType:
		// 建立私聊关系：是在用户点击发起聊天后触发
		conversationId := wuid.CombineId(in.SendId, in.RecvId)
		// 建立两者会话
		_, err := l.svcCtx.ConversationModel.FindOne(l.ctx, conversationId)
		if err != nil {
			if err == immodels.ErrNotFound {
				err = l.svcCtx.ConversationModel.Insert(l.ctx, &immodels.Conversation{
					ConversationId: conversationId,
					Msg:            &immodels.ChatLog{},
					ChatType:       constants.SingleChatType,
				})
				if err != nil {
					return nil, errors.Wrapf(xerr.NewDBErr(), "create conversation err %v, req %v", err, in)
				} else {
					return nil, errors.Wrapf(xerr.NewDBErr(), "find conversation err %v, req %v", err, in)
				}
			}
			err = l.setUpUserConversation(conversationId, in.SendId, constants.SingleChatType, true)
			if err != nil {
				return &res, errors.Wrapf(err, "set up user single conversation err %v, req %v", err, in)
			}
			// 接收者是被动与目标用户建立连接，因此理论上是不需要在会话列表里面展示，而是在用户发起聊天后展示
			err = l.setUpUserConversation(conversationId, in.SendId, constants.SingleChatType, false)
			if err != nil {
				return &res, errors.Wrapf(err, "set up user conversation err %v, req %v", err, in)
			}
		}
	case constants.GroupChatType:
		// 建立群聊：动作触发时在加群通过后
		// 用户加入群后应该在会话里显示群聊
		// 接收者就是群Id
		err := l.setUpUserConversation(in.RecvId, in.SendId, constants.GroupChatType, true)
		if err != nil {
			return &res, errors.Wrapf(err, "set up user conversation err %v, req %v", err, in)
		}
	}

	return &res, nil
}

func (l *SetUpUserConversationLogic) setUpUserConversation(conversationId, userId string, chatType constants.ChatType, isShow bool) error {
	// 发送者
	conversations, err := l.svcCtx.ConversationsModel.FindByUserId(l.ctx, userId)
	if err != nil {
		if err == immodels.ErrNotFound {
			conversations = &immodels.Conversations{
				ID:               primitive.NewObjectID(),
				UserId:           userId,
				ConversationList: make(map[string]*immodels.Conversation),
			}
		} else {
			return err
		}
	}

	// 更新会话记录
	if _, ok := conversations.ConversationList[conversationId]; ok {
		// 存在
		return nil
	}
	// 需要建立
	conversations.ConversationList[conversationId] = &immodels.Conversation{
		ConversationId: conversationId,
		ChatType:       chatType,
		IsShow:         isShow,
	}

	// 存在就更新，不存在则修改
	err = l.svcCtx.ConversationsModel.Update(l.ctx, conversations)
	return err
}

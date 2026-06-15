package logic

import (
	"IM/apps/social/socialmodels"
	"IM/pkg/constants"
	"IM/pkg/xerr"
	"context"

	"IM/apps/social/rpc/internal/svc"
	"IM/apps/social/rpc/social"

	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var (
	ErrFriendReqBeforePass   = xerr.NewMsg("好友申请并通过")
	ErrFriendReqBeforeRefuse = xerr.NewMsg("好友申请被拒绝")
)

type FriendPutInHandleLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewFriendPutInHandleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FriendPutInHandleLogic {
	return &FriendPutInHandleLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *FriendPutInHandleLogic) FriendPutInHandle(in *social.FriendPutInHandleReq) (*social.FriendPutInHandleResp, error) {
	// todo: add your logic here and delete this line

	// 获取好友申请记录
	friendReq, err := l.svcCtx.FriendRequestsModel.FindOne(l.ctx, int64(in.FriendReqId))
	if err != nil {
		return nil, errors.Wrapf(xerr.NewDBErr(), "friend friendsRequest by friendReqId err %v req %v", err, in.FriendReqId)
	}
	// 验证是否有处理
	switch constants.HandlerResult(friendReq.HandleResult.Int64) {
	case constants.PassHandlerResult:
		return nil, errors.WithStack(ErrFriendReqBeforePass)
	case constants.RefuseHandlerResult:
		return nil, errors.WithStack(ErrFriendReqBeforeRefuse)
	}

	friendReq.HandleResult.Int64 = int64(in.HandleResult)
	// 修改申请结果 -》 通过【建立两条好友关系记录】-》事务
	err = l.svcCtx.FriendRequestsModel.Trans(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		if err := l.svcCtx.FriendRequestsModel.Update(l.ctx, session, friendReq); err != nil {
			return errors.Wrapf(xerr.NewDBErr(), "update friend request err %v req %v", err, friendReq)
		}
		if constants.HandlerResult(in.HandleResult) != constants.PassHandlerResult {
			return nil
		}

		friends := []*socialmodels.Friends{
			{
				UserId:    friendReq.UserId,
				FriendUid: friendReq.ReqUid,
			}, {
				UserId:    friendReq.UserId,
				FriendUid: friendReq.ReqUid,
			},
		}

		_, err := l.svcCtx.FriendsModel.Inserts(l.ctx, session, friends...)
		if err != nil {
			return errors.Wrapf(xerr.NewDBErr(), "insert friends err %v req %v", err, friends)
		}
		return nil
	})
	return &social.FriendPutInHandleResp{}, nil
}

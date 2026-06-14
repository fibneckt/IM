package logic

import (
	"IM/pkg/constants"
	"context"
	"database/sql"
	"time"

	"IM/apps/social/rpc/internal/svc"
	"IM/apps/social/rpc/social"
	"IM/apps/social/socialmodels"
	"IM/pkg/xerr"

	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
)

type FriendPutInLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewFriendPutInLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FriendPutInLogic {
	return &FriendPutInLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 好友业务
func (l *FriendPutInLogic) FriendPutIn(in *social.FriendPutInReq) (*social.FriendPutInResp, error) {
	// todo: add your logic here and delete this line

	// 申请人是否与目标是好友关系
	// uid, fid
	friends, err := l.svcCtx.FriendsModel.FindByUidAndFid(l.ctx, in.UserId, in.ReqUid)
	if err != nil && err != socialmodels.ErrNotFound {
		return nil, errors.Wrapf(xerr.NewDBErr(), "find friends by uid and fid err %v req %v", err, in)
	}
	if friends != nil {
		return &social.FriendPutInResp{}, nil
	}

	// 是否已经有过申请，申请不成功，没有完成
	friendReqs, err := l.svcCtx.FriendRequestsModel.FindByReqUidAndUserId(l.ctx, in.ReqUid, in.UserId)
	if err != nil && err != socialmodels.ErrNotFound {
		return nil, errors.Wrapf(xerr.NewDBErr(), "find friend requests by req uid and user id err %v req %v", err, in)
	}

	if friendReqs != nil {
		return &social.FriendPutInResp{}, err
	}
	// 创建申请记录
	_, err = l.svcCtx.FriendRequestsModel.Insert(l.ctx, &socialmodels.FriendRequests{
		UserId: in.UserId,
		ReqUid: in.ReqUid,
		ReqMsg: sql.NullString{
			Valid:  true,
			String: in.ReqMsg,
		},
		ReqTime: time.Unix(in.ReqTime, 0),
		HandleResult: sql.NullInt64{
			Int64: int64(constants.NoHandlerResult),
			Valid: true,
		},
	})

	if err != nil {
		return nil, errors.Wrapf(xerr.NewDBErr(), "insert friend requests err %v req %v", err, in)
	}
	return &social.FriendPutInResp{}, nil
}

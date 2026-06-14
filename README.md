# IM 通讯系统

## 用户服务

### 注册功能

```proto
message RegisterReq {
  string phone = 1;
  string nickname = 2;
  string password = 3;
  string avatar = 4;
  int32 sex = 5;
}

message RegisterResp {
  string Token = 1;
  int64 expire = 2;
}
```

```go
func (l *RegisterLogic) Register(in *user.RegisterReq) (*user.RegisterResp, error) {
	// todo: add your logic here and delete this line

	// 1. 验证用户名是否注册，根据手机号码验证
	userEntity, err := l.svcCtx.FindByPhone(l.ctx, in.Phone)
	if err != nil && err != models.ErrNotFound {
		return nil, err
	}

	if userEntity != nil {
		return nil, ErrPhoneIsRegister
	}

	// 1. 定义用户数据
	userEntity = &models.User{
		Id:       wuid.GenUid(l.svcCtx.Config.Mysql.DataSource),
		Avatar:   in.Avatar,
		Nickname: in.Nickname,
		Phone:    in.Phone,
		Sex: sql.NullInt64{
			Int64: int64(in.Sex),
			Valid: true,
		},
	}

	if len(in.Password) > 0 {
		genPassword, err := encrypt.GenPasswordHash([]byte(in.Password))
		if err != nil {
			return nil, err
		}
		userEntity.Password = sql.NullString{
			String: string(genPassword),
			Valid:  true,
		}
	}

	// 2. 写入数据库
	_, err = l.svcCtx.Insert(l.ctx, userEntity)
	if err != nil {
		return nil, err
	}

	// 3. 生成token
	now := time.Now().Unix()
	token, err := ctxdata.GetJwtToken(l.svcCtx.Config.Jwt.AccessSecret, now, l.svcCtx.Config.Jwt.AccessExpire, userEntity.Id)

	if err != nil {
		return nil, err
	}

	return &user.RegisterResp{
		Token:  token,
		Expire: now + l.svcCtx.Config.Jwt.AccessExpire,
	}, nil
}
```
### 登入功能

```proto
message LoginReq {
  string phone = 1;
  string password = 2;
}

message LoginResp {
  string Token = 1;
  int64 expire = 2;
}
```

```go
func (l *LoginLogic) Login(in *user.LoginReq) (*user.LoginResp, error) {
	// todo: add your logic here and delete this line

	// 1. 验证用户是否注册，根据手机号验证
	userEntity, err := l.svcCtx.UserModel.FindByPhone(l.ctx, in.Phone)
	if err != nil {
		if err == models.ErrNotFound {
			return nil, errors.WithStack(ErrPhoneNotRegister)
		}
		return nil, errors.Wrapf(xerr.NewDBErr(), "find user by phone err %v , req %v", err, in.Phone)
	}

	// 2. 密码验证
	if !encrypt.ValidatePassword(in.Password, userEntity.Password.String) {
		return nil, errors.WithStack(ErrUserPwdError)
	}

	// 3. 生成token
	now := time.Now().Unix()
	token, err := ctxdata.GetJwtToken(l.svcCtx.Config.Jwt.AccessSecret, now, l.svcCtx.Config.Jwt.AccessExpire, userEntity.Id)
	if err != nil {
		return nil, errors.Wrapf(xerr.NewDBErr(), "ctxdata get jwt token err %v ", err)
	}

	return &user.LoginResp{
		Token:  token,
		Expire: now + l.svcCtx.Config.Jwt.AccessExpire,
	}, nil
}
```

### 获取用户信息功能

```proto
message GetUserInfoReq {
  string id = 1;
}

message GetUserInfoResp {
  UserEntity user = 1;
}
```

```go
func (l *GetUserInfoLogic) GetUserInfo(in *user.GetUserInfoReq) (*user.GetUserInfoResp, error) {
	// todo: add your logic here and delete this line

	userEntity, err := l.svcCtx.UserModel.FindOne(l.ctx, in.Id)
	if err != nil {
		if err == models.ErrNotFound {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	var resp user.UserEntity
	copier.Copy(&resp, userEntity)

	return &user.GetUserInfoResp{
		User: &resp,
	}, nil
}
```

### 寻找用户功能

```proto
message FindUserReq {
  string name = 1;
  string phone = 2;
  repeated string ids = 3;
}

message FindUserResp {
  repeated UserEntity user = 1;
}
```

```go
func (l *FindUserLogic) FindUser(in *user.FindUserReq) (*user.FindUserResp, error) {
	// todo: add your logic here and delete this line

	var (
		userEntitys []*models.User
		err         error
	)

	if in.Phone != "" {
		userEntity, err := l.svcCtx.UserModel.FindByPhone(l.ctx, in.Phone)
		if err == nil {
			userEntitys = append(userEntitys, userEntity)
		}
	} else if in.Name != "" {
		userEntitys, err = l.svcCtx.UserModel.ListByName(l.ctx, in.Name)
	} else if len(in.Ids) > 0 {
		userEntitys, err = l.svcCtx.UserModel.ListByIds(l.ctx, in.Ids)
	}

	if err != nil {
		return nil, err
	}

	var resp []*user.UserEntity
	copier.Copy(&resp, userEntitys)

	return &user.FindUserResp{
		User: resp,
	}, nil
}
```

## 社区功能

### 好友申请

```proto
message FriendPutInReq {
  string userId = 2;
  string reqUid = 3;
  string reqMsg = 4;
  int64 reqTime = 5;
}

message FriendPutInResp {}
```

```go
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
```

### 好友申请处理

```proto
message FriendPutInHandleReq {
  int32 friendReqId = 1;
  string userId = 2;
  int32 handleResult = 3; // 处理结果
}

message FriendPutInHandleResp {}
```

```go
func (l *FriendPutInHandleLogic) FriendPutInHandle(in *social.FriendPutInHandleReq) (*social.FriendPutInHandleResp, error) {
	// todo: add your logic here and delete this line

	// 获取好友申请记录
	friendReq, err := l.svcCtx.FriendRequestsModel.FindOne(l.ctx, uint64(in.FriendReqId))
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
```


### 好友列表处理

```proto
message FriendPutInListReq {
  string userId = 1;
}

message FriendPutInListResp {
  repeated FriendRequests list = 1;
}
```

```go
func (l *FriendListLogic) FriendList(in *social.FriendListReq) (*social.FriendListResp, error) {
	// todo: add your logic here and delete this line

	friendsList, err := l.svcCtx.FriendsModel.ListByUserId(l.ctx, in.UserId)
	if err != nil {
		return nil, errors.Wrapf(xerr.NewDBErr(), "list friend by uid err %v req %v", err, in.UserId)
	}

	var respList []*social.Friends
	copier.Copy(&respList, &friendsList)

	return &social.FriendListResp{
		List: respList,
	}, nil
}
```
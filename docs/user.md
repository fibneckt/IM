# 用户服务

## 业务

在用户模块中，主要是用户的基本信息为主，主要功能有
- 注册
- 登入
- 用户信息
- 用户查找

## 项目结构

- apps 应用目录，记录相关服务信息
    - user
        - api
        - rpc
    - im
- deploy 项目部署相关的信息，如部署的时候一些程序的配置，sql或dockfile(在这里dockfile暂时不用写，项目全部完成再慢慢写)
- pkg 项目的公共工具目录
- dockfile-compose.yml 同dockfile一样
- Makefile 项目编译脚本工具(同dockfile一样)

## 构建项目

根据业务需求，初步构建user的rpc服务方法以及服务方法提供的信息数据，并通过命令构建好user服务

rpc服务

```protobuf
//  ./apps/user/rpc/user.proto
syntax = "proto3";

option go_package = "./user";

// model
message UserEntity {
  string id = 1;      // 用户 id
  string avatar = 2;  // 用户头像
  string nickname = 3;// 昵称
  string phone = 4;   // 手机好吗
  string status = 5;  // 是否锁住
  int32 sex = 6;
}

// req and resp
message Request {
  string ping = 1;
}

message Response {
  string pong = 1;
}

message LoginReq {
  string phone = 1;
  string password = 2;
}

message LoginResp {
  string Token = 1;
  int64 expire = 2;
}

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

message GetUserInfoReq {
  string id = 1;
}

message GetUserInfoResp {
  UserEntity user = 1;
}

message FindUserReq {
  string name = 1;
  string phone = 2;
  repeated string ids = 3;
}

message FindUserResp {
  repeated UserEntity user = 1;
}

service User {
  rpc Ping(Request) returns (Response);

  rpc Login(LoginReq) returns (LoginResp);

  rpc Register(RegisterReq) returns (RegisterResp);

  rpc GetUserInfo(GetUserInfoReq) returns(GetUserInfoResp);

  rpc FindUser(FindUserReq) returns(FindUserResp);
}
```

命令

```shell
// ./apps/user/bin/exec.sh
goctl rpc protoc ./apps/user/rpc/user.proto \
  --go_out=./apps/user/rpc/ \
  --go-grpc_out=./apps/user/rpc/ \
  --zrpc_out=./apps/user/rpc/
```

### 代码实现

#### 用户数据结构

sql 表

```sql
CREATE TABLE `users` (
         `Id` varchar(24) COLLATE utf8mb4_unicode_ci  NOT NULL ,
         `avatar` varchar(191) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
         `nickname` varchar(24) COLLATE utf8mb4_unicode_ci NOT NULL,
         `phone` varchar(20) COLLATE utf8mb4_unicode_ci NOT NULL,
         `password` varchar(191) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
         `status` tinyint COLLATE utf8mb4_unicode_ci DEFAULT NULL,
         `sex` tinyint COLLATE utf8mb4_unicode_ci DEFAULT NULL,
         `created_at` timestamp NULL DEFAULT NULL,
         `updated_at` timestamp NULL DEFAULT NULL,
         PRIMARY KEY (`Id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
```

项目结构

- apps 
    - user
        - rpc
        - models
- deploy 
    - sql

命令

```shell
// ./apps/user/bin/exec.sh
goctl model mysql ddl -src="./deploy/sql/user.sql" -dir="./apps/user/models" -c
```

#### 前置准备

在用户服务中的主要业务功能是，先验证用户的手机号码是否存在，如果存在则不允许注册，其次是注册成功后在业务需下也会给用户提供对应的jwt token

配置

```yaml
# ./apps/user/rpc/etc/dev/user.yaml
Name: user.rpc
ListenOn: 0.0.0.0:10000
Etcd:
   Hosts:
   - 127.0.0.1:2379
   Key: user.rpc

Mysql:
  DataSource: root:@tcp(127.0.0.1:3306)/im?charset=utf8mb4&loc=UTC

Cache:
  # redis
   - Host: 127.0.0.1:6379
     Type: node
     Pass:

Jwt:
  AccessSecret: im.fibneckt
  AccessExpire: 8640000
```

```go
// ./apps/user/rpc/internal/config/config.go
type Config struct {
	zrpc.RpcServerConf

	Mysql struct {
		DataSource string
	}

	Cache cache.CacheConf

	Jwt struct {
		AccessSecret string
		AccessExpire int64
	}
}
```

在服务的核心对象中增加对user的引用

```go
// ./apps/user/rpc/internal/svc/servicecontext.go
type ServiceContext struct {
	Config config.Config
	*redis.Redis
	models.UserModel
}

func NewServiceContext(c config.Config) *ServiceContext {
    sqlConn := sqlx.NewMysql(c.Mysql.DataSource)
    
    return &ServiceContext{
        Config:    c,
        Redis:     redis.MustNewRedis(c.Redisx),
        UserModel: models.NewUserModel(sqlConn, c.Cache),
    }
}
```

需要给用户验证是否存在已经注册过的手机号码，所以在user的model中需要添加基于手机号查询的方式

```go
// ./apps/user/models/usermodel_gen.go
func (m *defaultUserModel) FindByPhone(ctx context.Context, phone string) (*User, error) {
    cacheKey := fmt.Sprintf("%s%v", cacheUserIdPrefix, phone)
	
    var resp User
	
    err := m.QueryRowCtx(ctx, &resp, cacheKey, func(ctx context.Context, conn sqlx.SqlConn, v any) error {
        query := fmt.Sprintf("select %s from %s where `phone` = ? limit 1", userRows, m.table)
        return conn.QueryRowCtx(ctx, v, query, phone)
    })
    switch err {
        case nil:
            return &resp, nil
        case sqlc.ErrNotFound:
            return nil, ErrNotFound
        default:
            return nil, err
    }
}
```

还需对用户进行加密处理以及在使用jwt的情况下仍然需获取到用户的uid信息，因此在程序中提供两个公共的包，至于pkg下以便于整个项目的调用

```go
// ./pkg/encrypt/hash.go
// 加密
func Md5(str []byte) string {
	h := md5.New()
	h.Write(str)
	return hex.EncodeToString(h.Sum(nil))
}

// hash加密
func GenPasswordHash(password []byte) ([]byte, error) {
	return bcrypt.GenerateFromPassword(password, bcrypt.DefaultCost)
}

// hash校验
func ValidatePasswordHash(password string, hashed string) bool {
	if err := bcrypt.CompareHashAndPassword([]byte(hashed), []byte(password)); err != nil {
		return false
	}
	return true
}
```

```go
// ./pkg/ctxdata/token.go
func GetJwtToken(secretKey string, iat, seconds int64, uid string) (string, error) {
	claims := make(jwt.MapClaims)
	claims["exp"] = iat + seconds
	claims["iat"] = iat
	claims[Identify] = uid

	token := jwt.New(jwt.SigningMethodHS256)
	token.Claims = claims

	return token.SignedString([]byte(secretKey))
}
```

```go
// ./pkg/ctxdata/data.go
func GetUid(ctx context.Context) string {
	if u, ok := ctx.Value(Identify).(string); ok {
		return u
	}
	return ""
}
```

关于id可以最简单的直接使用自增

但如果考虑后续数据量的增长过大的需拆分数据库的时候则需要细节思考

这里采用第三方库wuid来生成id

```go
var w *wuid.WUID

func Init(dsn string) {

	newDB := func() (*sql.DB, bool, error) {
		db, err := sql.Open("mysql", dsn)
		if err != nil {
			return nil, false, err
		}
		return db, true, nil
	}

	w = wuid.NewWUID("default", nil)
	_ = w.LoadH28FromMysql(newDB, "wuid")
}

func GenUid(dsn string) string {
	if w == nil {
		Init(dsn)
	}

	return fmt.Sprintf("%#016x", w.Next())
}
```

这个工具会记录创建的id位置存储方式可以是redis也可以用mysql

如下是mysql的数据表结构

```sql
CREATE TABLE `wuid` (
    `h` int(10) NOT NULL AUTO_INCREMENT,
    `x` tinyint(4) NOT NULL DEFAULT '0',
    PRIMARY KEY (`x`),
    UNIQUE KEY `h` (`h`)
) ENGINE=InnoDB AUTO_INCREMENT=0 DEFAULT CHARSET=latin1;
```

工具的特点是生成的数据具有以下特点

- WUID是一个通用唯一标识符生成器
- WUID比传统UUID快得多，每个WUID实例每秒可以生成1亿个唯一标识符
- 本质上，WUID按顺序生成64位整数，高28位从数据源加载，目前支持Redis、MySQL、MongoDB和Callback
- 只要所有WUID实例共享相同数据源或每组具有不同的section ID，就能保证唯一性
- 当低36位即将用完时，WUID会自动更新高28位
- WUID是线程安全的，并且无锁
- 支持混淆

#### 用户注册功能

```go
// ./apps/user/rpc/internal/logic/registerlogic.go
var (
	ErrPhoneIsRegister = errors.New("手机号已经注册过")
)

type RegisterLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRegisterLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RegisterLogic {
	return &RegisterLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *RegisterLogic) Register(in *user.RegisterReq) (*user.RegisterResp, error) {
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

#### 用户登录功能

```go
// ./apps/user/rpc/internal/logic/loginlogic.go
var (
	ErrPhoneNotRegister = xerr.New(xerr.SERVER_COMMON_ERROR, "手机号没有注册")
	ErrUserPwdError     = xerr.New(xerr.SERVER_COMMON_ERROR, "用户密码错误")
)

type LoginLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginLogic {
	return &LoginLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *LoginLogic) Login(in *user.LoginReq) (*user.LoginResp, error) {
	// 1. 验证用户是否注册，根据手机号验证
	userEntity, err := l.svcCtx.UserModel.FindByPhone(l.ctx, in.Phone)
	if err != nil {
		if err == models.ErrNotFound {
			return nil, errors.WithStack(ErrPhoneNotRegister)
		}
		return nil, errors.Wrapf(xerr.NewDBErr(), "find user by phone err %v , req %v", err, in.Phone)
	}

	// 2. 密码验证
	if !encrypt.ValidatePasswordHash(in.Password, userEntity.Password.String) {
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

#### 搜索用户

用户搜索的功能对于业务来说方式有很多种，因此需要做好区分

调整搜索的请求参数需求，在该方法中只使用一个条件进行处理而非多个条件一起

在用户模型中增加根据name以及ids查询方法

在查询的功能中，使用的是非缓存的方式进行查询，请求是直接走的数据库

```go
// ./apps/user/models/usermodel_gen.go
type (
	userModel interface {
		ListByName(ctx context.Context, name string) ([]*User, error)
		ListByIds(ctx context.Context, ids []string) ([]*User, error)
	}
)

func (m *defaultUserModel) ListByName(ctx context.Context, name string) ([]*User, error) {
    query := fmt.Sprintf("select %s from %s where `nickname` like ?", userRows, m.table)
    
    var resp []*User
    err := m.QueryRowsNoCacheCtx(ctx, &resp, query, fmt.Sprint(name, "%", name, "%"))
    switch err {
        case nil:
            return resp, nil
        default:
            return nil, err
    }
}

func (m *defaultUserModel) ListByIds(ctx context.Context, ids []string) ([]*User, error) {
    query := fmt.Sprintf("select %s from %s where `id` in ('%s') ", userRows, m.table)
    
    var resp []*User
    err := m.QueryRowsNoCacheCtx(ctx, &resp, query)
        switch err {
        case nil:
            return resp, nil
        default:
            return nil, err
    }
}
```

```go
// ./apps/user/rpc/internal/logic/finduserlogic.go
func NewFindUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FindUserLogic {
	return &FindUserLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *FindUserLogic) FindUser(in *user.FindUserReq) (*user.FindUserResp, error) {
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

#### 获取用户信息

主要是基于id的查询用户信息并返回用户的信息记录

```go
// ./apps/user/rpc/internal/logic/getuserinfologic.go
func NewGetUserInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserInfoLogic {
	return &GetUserInfoLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetUserInfoLogic) GetUserInfo(in *user.GetUserInfoReq) (*user.GetUserInfoResp, error) {
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

#### 错误处理

针对错误处理目前的方式是将系统错误信息直接返回给了调用者，从程序上着不是特别友好没有做异常的处理

在程序中对异常处理主要有异常信息，状态码及后续异常产生的日志信息，使用go-zero中提供对errors处理的方式

可以在服务中增加对错误的日志信息记录，利用grpc的拦截器实现

```go
// ./pkg/interceptor/rpcserver/LoginInterceptor.go
func LoginInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
	resp, err = handler(ctx, req)
	if err == nil {
		return resp, err
	}

	logx.WithContext(ctx).Errorf("[RPC SRV ERR] %v", err)

	causeErr := errors.Cause(err)
	if e, ok := causeErr.(*zerr.CodeMsg); ok {
		err = status.Error(codes.Code(e.Code), e.Msg)
	}
	return resp, err
}

```

然后在服务中引用

```go
func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)
	ctx := svc.NewServiceContext(c)

	if err := ctx.SetRootToken(); err != nil {
		panic(err)
	}

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		user.RegisterUserServer(grpcServer, server.NewUserServer(ctx))

		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	s.AddUnaryInterceptors(rpcserver.LoginInterceptor) //引用
	defer s.Stop()

	fmt.Printf("Starting rpc server at %s...\n", c.ListenOn)
	s.Start()
}
```

package svc

import (
	"IM/apps/user/models"
	"IM/apps/user/rpc/internal/config"
	"IM/pkg/constants"
	"IM/pkg/ctxdata"
	"context"
	"time"

	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ServiceContext struct {
	Config config.Config
	*redis.Redis
	models.UserModel
}

func (s *ServiceContext) ListByIds(ctx context.Context, ids []string) ([]*models.User, error) {
	return s.UserModel.ListByIds(ctx, ids)
}

func (s *ServiceContext) ListByName(ctx context.Context, name string) ([]*models.User, error) {
	return s.UserModel.ListByName(ctx, name)
}

func NewServiceContext(c config.Config) *ServiceContext {
	sqlConn := sqlx.NewMysql(c.Mysql.DataSource)

	return &ServiceContext{
		Config:    c,
		Redis:     redis.MustNewRedis(c.Redisx),
		UserModel: models.NewUserModel(sqlConn, c.Cache),
	}
}

func (svc *ServiceContext) SetRootToken() error {
	// 生成 jwt
	systemToken, err := ctxdata.GetJwtToken(svc.Config.Jwt.AccessSecret, time.Now().Unix(), 9999999, constants.REDIS_SYSTEM_ROOT_TOKEN)
	if err != nil {
		return err
	}
	// 写入到redis
	return svc.Redis.Set(constants.REDIS_SYSTEM_ROOT_TOKEN, systemToken)
}

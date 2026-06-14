package svc

import (
	"IM/apps/user/models"
	"IM/apps/user/rpc/internal/config"
	"context"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ServiceContext struct {
	Config config.Config

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
		Config: c,

		UserModel: models.NewUserModel(sqlConn, c.Cache),
	}
}

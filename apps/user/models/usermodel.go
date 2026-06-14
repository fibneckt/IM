package models

import (
	"context"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ UserModel = (*customUserModel)(nil)

type (
	// UserModel is an interface to be customized, add more methods here,
	// and implement the added methods in customUserModel.
	UserModel interface {
		userModel
		ListByIds(ctx context.Context, ids []string) ([]*User, error)
		ListByName(ctx context.Context, name string) ([]*User, error)
	}

	customUserModel struct {
		*defaultUserModel
	}
)

func (m *customUserModel) ListByIds(ctx context.Context, ids []string) ([]*User, error) {
	// TODO implement me
	panic("implement me")
}

func (m *customUserModel) ListByName(ctx context.Context, name string) ([]*User, error) {
	// TODO implement me
	panic("implement me")
}

// NewUserModel returns a model for the database table.
func NewUserModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) UserModel {
	return &customUserModel{
		defaultUserModel: newUserModel(conn, c, opts...),
	}
}

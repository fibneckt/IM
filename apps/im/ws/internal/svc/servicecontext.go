package svc

import (
	"IM/apps/im/immodels"
	"IM/apps/im/ws/internal/config"
)

type ServiceContext struct {
	Config config.Config

	immodels.ChatLogModel
}

func NewServiceContext(config config.Config) *ServiceContext {
	return &ServiceContext{
		Config: config,

		ChatLogModel: immodels.MustChatLogModel(config.Mongo.Url, config.Mongo.Db),
	}
}

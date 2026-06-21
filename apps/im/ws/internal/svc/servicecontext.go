package svc

import (
	"IM/apps/im/immodels"
	"IM/apps/im/ws/internal/config"
	"IM/apps/task/mq/mqclient"
)

type ServiceContext struct {
	Config config.Config

	immodels.ChatLogModel
	mqclient.MsgChatTransferClient
}

func NewServiceContext(config config.Config) *ServiceContext {
	return &ServiceContext{
		Config:                config,
		MsgChatTransferClient: mqclient.NewMsgChatTransferClient(config.MsgChatTransfer.Addrs, config.MsgChatTransfer.Topic),
		ChatLogModel:          immodels.MustChatLogModel(config.Mongo.Url, config.Mongo.Db),
	}
}

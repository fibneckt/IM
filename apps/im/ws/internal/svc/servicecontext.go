package svc

import "IM/apps/im/ws/internal/config"

type ServiceContext struct {
	Config config.Config
}

func NewServiceContext(config config.Config) *ServiceContext {
	return &ServiceContext{
		Config: config,
	}
}

package config

import "github.com/zeromicro/go-zero/core/service"

type Config struct {
	service.ServiceConf

	// 监听
	ListenOn string

	// JWT
	JwtAuth struct {
		AccessSecret string
	}
}

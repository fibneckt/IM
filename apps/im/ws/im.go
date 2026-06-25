package main

import (
	"IM/apps/im/ws/internal/config"
	"IM/apps/im/ws/internal/handler"
	"IM/apps/im/ws/internal/svc"
	"IM/apps/im/ws/websocket"
	"flag"
	"fmt"

	"github.com/zeromicro/go-zero/core/conf"
)

var configFile = flag.String("f", "etc/dev/im.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	if err := c.SetUp(); err != nil {
		panic(err)
	}

	ctx := svc.NewServiceContext(c)
	// 添加jwt 中间件
	srv := websocket.NewServer(c.ListenOn,
		websocket.WithServerAuthentication(handler.NewJwtAuth(ctx)),
		// 心跳机制最大时间限制
		// websocket.WithServerMaxConnectionIdle(10*time.Second),
		websocket.WithServerAck(websocket.OnlyAck),
	)
	defer srv.Stop()

	handler.RegisterHandlers(srv, ctx)

	fmt.Println("start websocket server at ", c.ListenOn, "....")
	srv.Start()

}

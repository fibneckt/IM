package main

import (
	"IM/apps/im/ws/internal/config"
	"IM/apps/im/ws/internal/handler"
	"IM/apps/im/ws/internal/svc"
	"IM/apps/im/ws/websocket"
	"flag"
	"fmt"
	"time"

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
		websocket.WithServerMaxConnectionIdle(3600*time.Second),
	)
	defer srv.Stop()

	handler.RegisterHandlers(srv, ctx)

	fmt.Println("start websocket server at ", c.ListenOn, "....")
	srv.Start()

}

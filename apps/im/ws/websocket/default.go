package websocket

import "time"

const (
	defaultMaxConnectionIdle = 10 * time.Second
	defaultAckTimeout        = 30 * time.Second
)

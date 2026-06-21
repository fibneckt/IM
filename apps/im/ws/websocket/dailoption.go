package websocket

import "net/http"

type DailOptions func(option *dailOption)

// 客户端操作
type dailOption struct {
	// 请求websocket的地址
	pattern string
	// 请求头参数会使用到
	header http.Header
}

func newDailOptions(opts ...DailOptions) dailOption {
	o := dailOption{
		pattern: "/ws",
		header:  nil,
	}
	for _, opt := range opts {
		opt(&o)
	}
	return o
}

func WithClientPatten(patten string) DailOptions {
	return func(opt *dailOption) {
		opt.pattern = patten
	}
}

func WithServerHeader(header http.Header) DailOptions {
	return func(opt *dailOption) {
		opt.header = header
	}
}

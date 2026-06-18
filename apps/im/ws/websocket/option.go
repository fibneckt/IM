package websocket

type ServerOptions func(opt *ServerOption)
type ServerOption struct {
	Authentication
	patten string
}

func newServerOptions(opts ...ServerOptions) ServerOption {
	o := ServerOption{
		Authentication: new(authentication),
		patten:         "/ws",
	}

	for _, opt := range opts {
		opt(&o)
	}
	return o
}

func WithServerAuthentication(auth Authentication) ServerOptions {
	return func(opt *ServerOption) {
		opt.Authentication = auth
	}
}

func WithServerPatten(patten string) ServerOptions {
	return func(opt *ServerOption) {
		opt.patten = patten
	}
}

package server

import "net/http"

type Option func(*options)

type options struct {
	handlers map[string]http.Handler
	port     string
}

func WithHandler(pattern string, handler http.Handler) Option {
	return func(o *options) {
		o.handlers[pattern] = handler
	}
}

func WithPort(port string) Option {
	return func(o *options) {
		o.port = port
	}
}

func initOptions() *options {
	return &options{
		handlers: make(map[string]http.Handler),
	}
}

package server

import "net/http"

type Option func(*Server)

func WithHandler(pattern string, handler http.Handler) Option {
	return func(s *Server) {
		s.handlers[pattern] = handler
	}
}

func WithPort(port string) Option {
	return func(s *Server) {
		if port != "" {
			s.port = port
		}
	}
}

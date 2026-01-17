package server

import (
	"log"
	"net/http"
)

type Server struct {
	port     string
	handlers map[string]http.Handler
}

func New(opts ...Option) *Server {
	s := &Server{
		port:     "8080",
		handlers: make(map[string]http.Handler),
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

func (s *Server) Run() error {
	for pattern, handler := range s.handlers {
		http.Handle(pattern, handler)
	}

	log.Println("Server started on port", s.port)
	return http.ListenAndServe(":"+s.port, nil)
}

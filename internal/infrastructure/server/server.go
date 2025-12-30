package server

import (
	"log"
	"net/http"
)

type Server struct {
	options []Option
}

func New(opts ...Option) *Server {
	return &Server{options: opts}
}

func (s *Server) Run() error {
	o := initOptions()
	for _, opt := range s.options {
		opt(o)
	}

	for pattern, handler := range o.handlers {
		http.Handle(pattern, handler)
	}

	log.Println("Server started on port", o.port)
	return http.ListenAndServe(":"+o.port, nil)
}

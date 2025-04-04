package server

import (
	"fmt"
	"net/http"

	"github.com/hse-telescope/gateway/internal/config"
)

type Provider interface{}

type Server struct {
	server   http.Server
	provider Provider
}

func New(conf config.Config, provider Provider) *Server {
	s := new(Server)
	s.server.Addr = fmt.Sprintf(":%d", conf.Port)
	s.server.Handler = s.setRouter()
	s.provider = provider
	return s
}

func (s *Server) setRouter() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /ping", s.pingHandler)
	mux.Handle("/auth", wrapHandlerFunc(s.authHandler))
	mux.Handle("/core", wrapHandlerFunc(s.coreHandler))
	return mux
}

func (s *Server) Start() error {
	return s.server.ListenAndServe()
}

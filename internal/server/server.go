package server

import (
	"fmt"
	"net/http"

	"github.com/hse-telescope/gateway/internal/config"
)

type provider interface {
	ProxyRequestAuth(req *http.Request) (*http.Response, error)
	ProxyRequestCore(req *http.Request) (*http.Response, error)
}

type Server struct {
	server   http.Server
	provider provider
}

func New(conf config.Config, provider provider) *Server {
	s := new(Server)
	s.server.Addr = fmt.Sprintf(":%d", conf.Port)
	s.server.Handler = s.setRouter()
	s.provider = provider
	return s
}

func (s *Server) setRouter() *http.ServeMux {
	mux := http.NewServeMux()
	mux.Handle("GET /ping", wrapHandlerFunc(s.pingHandler, addContext, addTracing))
	mux.Handle("/auth", wrapHandlerFunc(s.authHandler, addContext, addTracing, s.addAuthentification))
	mux.Handle("/core", wrapHandlerFunc(s.coreHandler, addContext, addTracing, s.addAuthentification))
	return mux
}

func (s *Server) Start() error {
	return s.server.ListenAndServe()
}

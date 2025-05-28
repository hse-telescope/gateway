package server

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hse-telescope/gateway/internal/config"
	"github.com/hse-telescope/gateway/internal/providers/token"
)

type Provider interface {
	ParseToken(token string) (token.UserInfo, bool)
}

type Client interface {
	Do(ctx context.Context, req *http.Request) (*http.Response, error)
	Host() string
}

type Server struct {
	server   http.Server
	provider Provider
	auth     Client
	core     Client
}

func New(conf config.Config, provider Provider, authClient Client, coreClient Client) *Server {
	s := new(Server)
	s.server.Addr = fmt.Sprintf(":%d", conf.Port)
	s.server.Handler = s.setRouter()
	s.provider = provider
	s.auth = authClient
	s.core = coreClient
	return s
}

func (s *Server) setRouter() *http.ServeMux {
	mux := http.NewServeMux()
	mux.Handle("GET /ping", wrapHandlerFunc(s.pingHandler, addContext, addTracing))
	mux.Handle("/auth/", wrapHandlerFunc(s.authHandler, addContext, addTracing, s.addAuthentification))
	mux.Handle("/core/", wrapHandlerFunc(s.coreHandler, addContext, addTracing, s.addAuthentification))
	return mux
}

func (s *Server) Start() error {
	return s.server.ListenAndServe()
}

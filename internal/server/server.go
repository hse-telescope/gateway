package server

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hse-telescope/gateway/internal/config"
	"github.com/hse-telescope/gateway/internal/providers/token"
	"github.com/hse-telescope/logger"
	"github.com/hse-telescope/tracer"
	"github.com/hse-telescope/utils/handlers"
	"github.com/prometheus/client_golang/prometheus/promhttp"
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
	mux.Handle("GET /ping", handlers.WrapHandlerFunc(s.pingHandler, addContext, tracer.AddTracingMiddleware, logger.AddLoggingMiddleware))
	mux.Handle("/metrics", promhttp.Handler())
	mux.Handle(authPath, handlers.WrapHandlerFunc(s.authHandler, addContext, tracer.AddTracingMiddleware, logger.AddLoggingMiddleware, s.addAuthentification))
	mux.Handle(corePath, handlers.WrapHandlerFunc(s.coreHandler, addContext, tracer.AddTracingMiddleware, logger.AddLoggingMiddleware, s.addAuthentification))
	return mux
}

func (s *Server) Start() error {
	return s.server.ListenAndServe()
}

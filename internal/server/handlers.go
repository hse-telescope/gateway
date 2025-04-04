package server

import (
	"context"
	"net/http"
	"time"
)

const (
	timeout = 1 * time.Second
)

func wrapHandlerFunc(handler http.HandlerFunc) http.Handler {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return addTracing(
		addContext(
			ctx,
			handler,
		),
	)
}

func addContext(ctx context.Context, handler http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			r = r.WithContext(ctx)
			handler.ServeHTTP(w, r)
		},
	)
}

func addTracing(handler http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			// TODO: add tracing
			handler.ServeHTTP(w, r)
		},
	)
}

func (s *Server) pingHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("pong"))
}

func (s *Server) authHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("auth"))
}

func (s *Server) coreHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("core"))
}

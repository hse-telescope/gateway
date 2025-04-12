package server

import (
	"context"
	"net/http"
	"time"
)

type middleware = func(http.Handler) http.Handler

const (
	timeout = 1 * time.Second
)

func wrapHandlerFunc(handlerFunc http.HandlerFunc, middlewares ...middleware) http.Handler {
	var handler http.Handler = handlerFunc
	for _, middleware := range middlewares {
		handler = middleware(handler)
	}
	return handler
}

func addContext(handler http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), timeout)
			defer cancel()

			r = r.WithContext(ctx)
			handler.ServeHTTP(w, r)
		},
	)
}

func addTracing(handler http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			// TODO: Add tracing
			handler.ServeHTTP(w, r)
		},
	)
}

func (s *Server) addAuthentification(handler http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			// TODO: Add authentification
			handler.ServeHTTP(w, r)
		},
	)
}

func (s *Server) pingHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("pong"))
}

func (s *Server) authHandler(w http.ResponseWriter, r *http.Request) {
	s.auth.Do(r.Context(), r)
}

func (s *Server) coreHandler(w http.ResponseWriter, r *http.Request) {
	s.core.Do(r.Context(), r)
}

package server

import (
	"context"
	"net/http"
	"strings"
	"time"
)

type middleware = func(http.Handler) http.Handler

const (
	timeout    = 1 * time.Second
	authHeader = "Authorization"

	authPath = "/auth"
	corePath = "/core"
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

func writeResp(w http.ResponseWriter, resp *http.Response) {
	w.WriteHeader(resp.StatusCode)
	body := make([]byte, 0)
	resp.Body.Read(body)
	w.Write(body)
}

func (s *Server) pingHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("pong"))
}

func (s *Server) authHandler(w http.ResponseWriter, r *http.Request) {
	r.URL.Path = strings.Replace(r.URL.Path, authPath, "/", 1)
	r.URL.Path = strings.ReplaceAll(r.URL.Path, "//", "/")
	resp, err := s.auth.Do(r.Context(), r)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal error"))
	}
	writeResp(w, resp)
}

func (s *Server) coreHandler(w http.ResponseWriter, r *http.Request) {
	r.URL.Path = strings.Replace(r.URL.Path, corePath, "/", 1)
	r.URL.Path = strings.ReplaceAll(r.URL.Path, "//", "/")

	// token := r.Header.Get(authHeader)
	// info, ok := s.provider.ParseToken(token)
	// if !ok {
	// 	w.WriteHeader(http.StatusUnauthorized)
	// 	w.Write([]byte("no valid token provided"))
	// 	return
	// }
	// r.Header.Add(authHeader, strconv.Itoa(info.UserID))

	resp, err := s.core.Do(r.Context(), r)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal error"))
	}
	writeResp(w, resp)
}

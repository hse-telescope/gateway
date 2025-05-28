package server

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
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
	defer resp.Body.Close()

	// Читаем всё тело
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Failed to read response body: %v", err)
		return
	}

	// Копируем заголовки
	for key, values := range resp.Header {
		w.Header()[key] = values
	}

	// Пишем ответ
	w.WriteHeader(resp.StatusCode)
	w.Write(body)

}

func (s *Server) pingHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("pong"))
}

func (s *Server) authHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(os.Stderr, "[GATEWAY] Incoming auth request: %s %s\n", r.Method, r.URL.Path)

	// Удаляем "/auth" из пути
	newPath := strings.TrimPrefix(r.URL.Path, authPath)
	if !strings.HasPrefix(newPath, "/") {
		newPath = "/" + newPath
	}

	req := r.Clone(r.Context())
	req.URL = &url.URL{
		Scheme:   "http",
		Host:     "auth:8080",
		Path:     newPath,
		RawQuery: r.URL.RawQuery,
	}
	req.RequestURI = ""

	fmt.Fprintf(os.Stderr, "[GATEWAY] Forwarding to: %s %s\n", req.Method, req.URL.String())

	// Делаем запрос
	resp, err := s.auth.Do(req.Context(), req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[GATEWAY] Auth request failed: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal error"))
		return
	}
	writeResp(w, resp)
}

func (s *Server) coreHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(os.Stderr, "[GATEWAY] Incoming request: %s %s\n", r.Method, r.URL.Path)
	newPath := strings.TrimPrefix(r.URL.Path, corePath)
	if newPath == "" {
		newPath = "/"
	}
	r.URL.Path = newPath

	targetURL := fmt.Sprintf("%s%s", s.core.Host(), r.URL.Path)
	fmt.Printf("[GATEWAY] Proxying to: %s\n", targetURL)

	proxyReq, err := http.NewRequest(r.Method, targetURL, r.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[GATEWAY] Proxy request error: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("proxy error: %v", err)))
		return
	}
	proxyReq.Header = r.Header
	resp, err := s.core.Do(r.Context(), proxyReq)
	if err != nil {
		fmt.Printf("[GATEWAY] Proxy DO error: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("proxy connection error: %v", err)))
		return
	}
	defer resp.Body.Close()

	for k, v := range resp.Header {
		w.Header()[k] = v
	}
	w.WriteHeader(resp.StatusCode)

	if _, err := io.Copy(w, resp.Body); err != nil {
		fmt.Fprintf(os.Stderr, "[GATEWAY] Response copy error: %v\n", err)
	}
}

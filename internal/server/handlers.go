package server

import (
	"context"
	"fmt"
	"io"
	"maps"
	"net/http"
	"strings"
	"time"

	"github.com/hse-telescope/logger"
)

const (
	timeout    = 1 * time.Second
	authHeader = "Authorization"

	authPath = "/api/auth"
	corePath = "/api/core"
)

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

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Failed to read response body: %v", err)
		return
	}

	maps.Copy(w.Header(), resp.Header)

	w.WriteHeader(resp.StatusCode)
	w.Write(body)
}

func (s *Server) pingHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("pong"))
}

func (s *Server) handler(w http.ResponseWriter, r *http.Request) {
	logger.Warn(r.Context(), "started handler")
	switch {
	case strings.Contains(r.URL.Path, authPath):
		s.authHandler(w, r)
	case strings.Contains(r.URL.Path, corePath):
		s.coreHandler(w, r)
	default:
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("failed to locate the page"))
	}
}

func (s *Server) authHandler(w http.ResponseWriter, r *http.Request) {
	logger.Debug(r.Context(), "[GATEWAY] Incoming auth request", "method", r.Method, "path", r.URL.Path)

	newPath := strings.TrimPrefix(r.URL.Path, corePath)
	if newPath == "" {
		newPath = "/"
	}
	r.URL.Path = newPath

	targetURL := fmt.Sprintf("%s%s", s.auth.Host(), r.URL.Path)

	proxyReq, err := http.NewRequest(r.Method, targetURL, r.Body)
	if err != nil {
		logger.Error(r.Context(), "[GATEWAY] Proxy request error", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("proxy error: %v", err)))
		return
	}

	logger.Error(r.Context(), "[GATEWAY] Forwarding", "method", proxyReq.Method, "url", proxyReq.URL.String())

	// Делаем запрос
	resp, err := s.auth.Do(r.Context(), proxyReq)
	if err != nil {
		logger.Error(r.Context(), "[GATEWAY] auth request error", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal error"))
		return
	}
	writeResp(w, resp)
}

func (s *Server) coreHandler(w http.ResponseWriter, r *http.Request) {
	logger.Debug(r.Context(), "[GATEWAY] Incoming request", "method", r.Method, "path", r.URL.Path)
	newPath := strings.TrimPrefix(r.URL.Path, corePath)
	if newPath == "" {
		newPath = "/"
	}
	r.URL.Path = newPath

	targetURL := fmt.Sprintf("%s%s", s.core.Host(), r.URL.Path)
	logger.Debug(r.Context(), "[GATEWAY] Incoming request", "method", r.Method, "path", r.URL.Path)
	logger.Debug(r.Context(), "[GATEWAY] Proxying to url", "url", targetURL)

	proxyReq, err := http.NewRequest(r.Method, targetURL, r.Body)
	if err != nil {
		logger.Error(r.Context(), "[GATEWAY] Proxy request error", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("proxy error: %v", err)))
		return
	}
	proxyReq.Header = r.Header
	resp, err := s.core.Do(r.Context(), proxyReq)
	if err != nil {
		logger.Error(r.Context(), "[GATEWAY] Proxy do error", "error", err)
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
		logger.Error(r.Context(), "[GATEWAY] Proxy response copy error", "error", err)
	}
}

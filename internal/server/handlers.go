package server

import (
	"context"
	"fmt"
	"io"
	"maps"
	"net/http"
	"os"
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
	fmt.Fprintf(os.Stderr, "[GATEWAY] Incoming auth request: %s %s\n", r.Method, r.URL.Path)

	newPath := strings.TrimPrefix(r.URL.Path, corePath)
	if newPath == "" {
		newPath = "/"
	}
	r.URL.Path = newPath

	targetURL := fmt.Sprintf("%s%s", s.auth.Host(), r.URL.Path)

	proxyReq, err := http.NewRequest(r.Method, targetURL, r.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[GATEWAY] Proxy request error: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("proxy error: %v", err)))
		return
	}

	fmt.Fprintf(os.Stderr, "[GATEWAY] Forwarding to: %s %s\n", proxyReq.Method, proxyReq.URL.String())

	// Делаем запрос
	resp, err := s.auth.Do(r.Context(), proxyReq)
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

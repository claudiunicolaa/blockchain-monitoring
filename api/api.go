package api

import (
	"context"
	"fmt"
	"net/http"
)

type HTTPServer struct {
	server *http.Server
	// service
}

func New(port string, handler http.Handler) *HTTPServer {
	addr := fmt.Sprintf(":%s", port)
	return &HTTPServer{
		server: &http.Server{Addr: addr, Handler: handler},
	}
}

func (s *HTTPServer) ListenAndServe(ctx context.Context) error {
	return s.server.ListenAndServe()
}

func (s *HTTPServer) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

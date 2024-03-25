package internalhttp

import (
	"net/http"
	"time"
)

type ResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (s *Server) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rw := &ResponseWriter{w, 0}
		start := time.Now()
		next.ServeHTTP(rw, r)
		duration := time.Since(start)
		s.logger.LogHTTPRequest(r, duration, rw.statusCode)
	})
}

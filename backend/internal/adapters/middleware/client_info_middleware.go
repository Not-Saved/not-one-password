package middleware

import (
	"context"
	"net"
	"net/http"
	"strings"
)

type contextKey string

const (
	IpContextKey        contextKey = "clientIP"
	UserAgentContextKey contextKey = "userAgent"
)

func (m *Middleware) ClientInfoMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := extractIP(r)
		ua := r.UserAgent()

		ctx := context.WithValue(r.Context(), IpContextKey, ip)
		ctx = context.WithValue(ctx, UserAgentContextKey, ua)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func extractIP(r *http.Request) string {
	// If behind reverse proxy
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		// First IP in list
		parts := strings.Split(forwarded, ",")
		return strings.TrimSpace(parts[0])
	}

	if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
		return realIP
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}

	return host
}

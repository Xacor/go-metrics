package middleware

import (
	"net"
	"net/http"
)

func WithCheckSubnet(trustedNet *net.IPNet) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cIP := net.ParseIP(r.Header.Get("X-Real-IP"))
			if !trustedNet.Contains(cIP) {
				w.WriteHeader(http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}

}

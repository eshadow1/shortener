package handler

import (
	"net"
	"net/http"

	"github.com/eshadow1/shortener/internal/configs"
)

// TrustedSubnetMiddleware проверяет, входит ли IP клиента в доверенную подсеть
func TrustedSubnetMiddleware(cfg *configs.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if cfg.TrustedSubnet == "" {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			_, ipNet, err := net.ParseCIDR(cfg.TrustedSubnet)
			if err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			ipStr := r.Header.Get("X-Real-IP")
			if ipStr == "" {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			clientIP := net.ParseIP(ipStr)
			if clientIP == nil {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			if !ipNet.Contains(clientIP) {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

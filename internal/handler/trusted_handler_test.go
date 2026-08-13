package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/eshadow1/shortener/internal/configs"
	"github.com/stretchr/testify/assert"
)

func TestTrustedSubnetMiddleware(t *testing.T) {
	cfg := &configs.Config{}

	mockHandler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte("OK"))
		if err != nil {
			return
		}
	})

	tests := []struct {
		name           string
		trustedSubnet  string
		xRealIP        string
		expectedStatus int
	}{
		{
			name:           "success_trusted_subnet",
			trustedSubnet:  "10.0.0.0/16",
			xRealIP:        "10.0.255.255",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "empty_trusted_subnet",
			trustedSubnet:  "",
			xRealIP:        "192.168.1.100",
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "valid_ip_outside_trusted_subnet",
			trustedSubnet:  "192.168.1.0/24",
			xRealIP:        "10.0.0.1",
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "no_X-Real-IP_header",
			trustedSubnet:  "192.168.1.0/24",
			xRealIP:        "",
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "invalid_ip_in_X-Real-IP_header",
			trustedSubnet:  "192.168.1.0/24",
			xRealIP:        "invalid-ip",
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "invalid_CIDR_trusted_subnet",
			trustedSubnet:  "invalid-cidr",
			xRealIP:        "192.168.1.50",
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cfg.TrustedSubnet = tc.trustedSubnet
			middleware := TrustedSubnetMiddleware(cfg)

			trustedHandler := middleware(mockHandler)

			req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/internal/stats", http.NoBody)
			req.Header.Set("X-Real-IP", tc.xRealIP)
			rr := httptest.NewRecorder()

			trustedHandler.ServeHTTP(rr, req)

			assert.Equal(t, tc.expectedStatus, rr.Code)
		})
	}
}

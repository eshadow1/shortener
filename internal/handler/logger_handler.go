package handler

import (
	"fmt"
	"net/http"
	"time"

	"github.com/eshadow1/shortener/internal/loggers"
)

func LoggerMiddleware() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			lrw := loggers.NewLoggingResponseWriter(w)

			next.ServeHTTP(lrw, r)

			duration := time.Since(start)

			loggers.Log.Infow("http_request",
				"method", r.Method,
				"uri", r.RequestURI,
				"duration", duration.String(),
				"status", lrw.ResponseData.Status,
				"size", fmt.Sprintf("%d bytes", lrw.ResponseData.Size),
			)
		})
	}
}

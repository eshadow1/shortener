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

			lrw := newLoggingResponseWriter(w)

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

type (
	responseData struct {
		Status int
		Size   int
	}

	loggingResponseWriter struct {
		http.ResponseWriter
		ResponseData *responseData
	}
)

func newLoggingResponseWriter(w http.ResponseWriter) *loggingResponseWriter {
	return &loggingResponseWriter{
		ResponseWriter: w,
		ResponseData:   &responseData{},
	}
}

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.ResponseData.Size += size
	return size, err
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.ResponseData.Status = statusCode
}

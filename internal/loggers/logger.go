package loggers

import (
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"
)

var Log *zap.SugaredLogger

func CreateLogger(level string) error {
	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return err
	}
	cfg := zap.NewProductionConfig()
	cfg.Level = lvl
	zl, err := cfg.Build()
	if err != nil {
		return err
	}
	Log = zl.Sugar()
	return nil
}

type (
	responseData struct {
		Status int
		Size   int
	}

	loggingResponseWriter struct {
		http.ResponseWriter
		responseData *responseData
	}
)

func NewLoggingResponseWriter(w http.ResponseWriter) *loggingResponseWriter {
	return &loggingResponseWriter{
		ResponseWriter: w,
		responseData:   &responseData{},
	}
}

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.Size += size
	return size, err
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.Status = statusCode
}

func RequestLogger() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			lrw := NewLoggingResponseWriter(w)

			next.ServeHTTP(lrw, r)

			duration := time.Since(start)

			Log.Infow("http_request",
				"method", r.Method,
				"uri", r.RequestURI,
				"duration", duration.String(),
				"status", lrw.responseData.Status,
				"size", fmt.Sprintf("%d bytes", lrw.responseData.Size),
			)
		})
	}
}

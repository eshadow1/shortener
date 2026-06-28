package handler

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/eshadow1/shortener/internal/audit"
	"github.com/eshadow1/shortener/internal/configs"
	"github.com/eshadow1/shortener/internal/loggers"
	"github.com/eshadow1/shortener/internal/model"
	"github.com/eshadow1/shortener/internal/repository"
	"github.com/eshadow1/shortener/internal/service"
	"github.com/go-chi/chi/v5"
)

const (
	defaultDB                  = "postgres://postgres:postgres@localhost/shorten?sslmode=disable"
	defaultAddr                = "localhost:8080"
	defaultBaseURL             = "http://localhost:8080"
	defaultLevelLog            = "info"
	defaultBufferSizeChan      = 100
	defaultBatchSize           = 10
	defaultFlushIntervalSecond = 15 * time.Second
	defaultUUID                = "test-test"
)

func routeInit() *chi.Mux {
	cfg := configs.NewConfig()

	cfg.Log.Level = defaultLevelLog
	cfg.Addr = defaultAddr
	cfg.BaseURL = defaultBaseURL

	cfg.Storage.PathDB = defaultDB

	cfg.Service.BatchSize = defaultBatchSize
	cfg.Service.FlushInterval = defaultFlushIntervalSecond
	cfg.Service.BufferSizeChan = defaultBufferSizeChan

	errCreateLog := loggers.CreateLogger(cfg.Log.Level)
	if errCreateLog != nil {
		fmt.Println("Error creating logger:", errCreateLog)
		return nil
	}

	var r service.Repository
	var rc service.RepoChecker
	if cfg.Storage.PathDB != "" {
		pdb, errCreate := repository.NewPostgreSQLRepository(cfg.Storage)
		if errCreate != nil {
			loggers.Log.Errorf("error creating connection db: %v", errCreate)
			return nil
		}
		r = pdb
		rc = pdb
	} else {
		r = repository.NewMemoryRepository(cfg.Storage.Path)
	}

	a := service.NewAuditBroker()

	if af := audit.NewFileObserver(cfg.Audit.File); af != nil {
		a.Register(af)
	}

	if ar := audit.NewRemoteObserver(cfg.Audit.URL); ar != nil {
		a.Register(ar)
	}

	s := service.NewShortenerService(r, cfg.Service)

	c := service.NewCheckerService(rc)
	h := NewHandler(cfg, s, c)

	rs := InitRouter(cfg, h, a)
	return rs
}

func BenchmarkRouter_PostCreate(b *testing.B) {
	mux := routeInit()

	body := `https://practicum.yandex.ru/test`

	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		req := httptest.NewRequestWithContext(b.Context(), http.MethodPost, "/", strings.NewReader(body))
		ctx := context.WithValue(req.Context(), model.UserIDContextKey, defaultUUID)
		*req = *req.WithContext(ctx)
		req.Header.Set("Content-Type", "text/plain")
		rr := httptest.NewRecorder()
		mux.ServeHTTP(rr, req)
	}
}

func BenchmarkRouter_Get(b *testing.B) {
	id := "e742b70d"

	mux := routeInit()

	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		req := httptest.NewRequestWithContext(b.Context(), http.MethodGet, "/"+id, http.NoBody)
		ctx := context.WithValue(req.Context(), model.UserIDContextKey, defaultUUID)
		*req = *req.WithContext(ctx)
		rr := httptest.NewRecorder()
		mux.ServeHTTP(rr, req)
	}
}

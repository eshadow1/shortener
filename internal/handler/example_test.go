package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/eshadow1/shortener/internal/audit"
	"github.com/eshadow1/shortener/internal/configs"
	"github.com/eshadow1/shortener/internal/loggers"
	"github.com/eshadow1/shortener/internal/model"
	"github.com/eshadow1/shortener/internal/repository"
	"github.com/eshadow1/shortener/internal/service"
	"github.com/go-chi/chi/v5"
)

func routeInitMemory() *chi.Mux {
	cfg := configs.NewConfig()

	cfg.Log.Level = defaultLevelLog
	cfg.Addr = defaultAddr
	cfg.BaseURL = defaultBaseURL

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

// Example_postCreate демонстрирует успешное создание короткого URL через POST с текстовым телом.
func Example_postCreate() {
	mux := routeInitMemory()

	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/", strings.NewReader(defaultBody))
	ctx := context.WithValue(req.Context(), model.UserIDContextKey, defaultUUID)
	*req = *req.WithContext(ctx)
	req.Header.Set("Content-Type", "text/plain")
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	fmt.Printf("body response: %s.\n", rr.Body.String())

	// Output:
	// body response: http://localhost:8080/e742b70d.
}

// Example_getOrigin демонстрирует успешный редирект с короткого URL на оригинальный.
func Example_getOrigin() {
	id := "e742b70d"

	mux := routeInitMemory()

	reqInit := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/", strings.NewReader(defaultBody))
	ctxInit := context.WithValue(reqInit.Context(), model.UserIDContextKey, defaultUUID)
	*reqInit = *reqInit.WithContext(ctxInit)
	reqInit.Header.Set("Content-Type", "text/plain")
	rrInit := httptest.NewRecorder()
	mux.ServeHTTP(rrInit, reqInit)

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/"+id, http.NoBody)
	ctx := context.WithValue(req.Context(), model.UserIDContextKey, defaultUUID)
	*req = *req.WithContext(ctx)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	fmt.Printf("Status code: %d.\n", rr.Code)

	// Output:
	// Status code: 307.
}

// Example_getCheckDB демонстрирует неудачную проверку доступности базы данных.
func Example_getCheckDB() {
	mux := routeInitMemory()

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/ping", http.NoBody)
	ctx := context.WithValue(req.Context(), model.UserIDContextKey, defaultUUID)
	*req = *req.WithContext(ctx)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	fmt.Printf("Status code: %d.\n", rr.Code)

	// Output:
	// Status code: 500.
}

// Example_postShorten демонстрирует успешное создание короткого URL через POST с JSON-телом.
func Example_postShorten() {
	mux := routeInitMemory()

	reqBody := model.OriginalInfo{OriginalURL: "https://practicum.yandex.ru/test2"}
	jsonBody, _ := json.Marshal(reqBody)
	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/api/shorten", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(req.Context(), model.UserIDContextKey, defaultUUID)
	*req = *req.WithContext(ctx)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	fmt.Printf("body response: %s.\n", rr.Body.String())

	// Output:
	// body response: {"result":"http://localhost:8080/2ed0f6f7"}
	// .
}

// Example_postShortenBatch демонстрирует успешное пакетное создание коротких URL.
func Example_postShortenBatch() {
	mux := routeInitMemory()

	reqBody := []model.OriginalInfo{
		{OriginalURL: "https://practicum.yandex.ru/test3", CorrelationID: "1"},
		{OriginalURL: "https://practicum.yandex.ru/test4", CorrelationID: "2"},
	}
	jsonBody, _ := json.Marshal(reqBody)
	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/api/shorten/batch", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(req.Context(), model.UserIDContextKey, defaultUUID)
	*req = *req.WithContext(ctx)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	fmt.Printf("Status code: %d.\n", rr.Code)
	fmt.Printf("body response: %s.\n", rr.Body.String())

	// Output:
	// Status code: 201.
	// body response: [{"short_url":"http://localhost:8080/12a829cc","correlation_id":"1"},{"short_url":"http://localhost:8080/5f861225","correlation_id":"2"}]
	// .
}

// Example_getUserURLs демонстрирует успешное получение списка URL пользователя.
func Example_getUserURLs() {
	mux := routeInitMemory()

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/user/urls", http.NoBody)
	ctx := context.WithValue(req.Context(), model.UserIDContextKey, defaultUUID)
	*req = *req.WithContext(ctx)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	fmt.Printf("Status code: %d.\n", rr.Code)
	fmt.Printf("body response: %s.\n", rr.Body.String())

	// Output:
	// Status code: 204.
	// body response: .
}

// Example_deleteUserURLs демонстрирует успешное массовое удаление URL пользователя.
func Example_deleteUserURLs() {
	mux := routeInitMemory()

	urlsToDelete := []string{"e742b70d"}
	jsonBody, _ := json.Marshal(urlsToDelete)
	req := httptest.NewRequestWithContext(context.Background(), http.MethodDelete, "/api/user/urls", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(req.Context(), model.UserIDContextKey, defaultUUID)
	*req = *req.WithContext(ctx)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	fmt.Printf("Status code: %d.\n", rr.Code)
	fmt.Printf("body response: %s.\n", rr.Body.String())

	// Output:
	// Status code: 202.
	// body response: Accepted.
}

package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/eshadow1/shortener/internal/model"
)

// Example_postCreate демонстрирует успешное создание короткого URL через POST с текстовым телом.
func Example_postCreate() {
	mux := routeInit()

	body := `https://practicum.yandex.ru/test`

	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/", strings.NewReader(body))
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

	mux := routeInit()

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/"+id, http.NoBody)
	ctx := context.WithValue(req.Context(), model.UserIDContextKey, defaultUUID)
	*req = *req.WithContext(ctx)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	fmt.Printf("Status code: %d.\n", rr.Code)

	// Output:
	// Status code: 307.
}

// Example_getCheckDB демонстрирует успешную проверку доступности базы данных.
func Example_getCheckDB() {
	mux := routeInit()

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/ping", http.NoBody)
	ctx := context.WithValue(req.Context(), model.UserIDContextKey, defaultUUID)
	*req = *req.WithContext(ctx)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	fmt.Printf("Status code: %d.\n", rr.Code)
	fmt.Printf("body response: %s.\n", rr.Body.String())

	// Output:
	// Status code: 200.
	// body response: OK.
}

// Example_postShorten демонстрирует успешное создание короткого URL через POST с JSON-телом.
func Example_postShorten() {
	mux := routeInit()

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
	mux := routeInit()

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
}

// Example_getUserURLs демонстрирует успешное получение списка URL пользователя.
func Example_getUserURLs() {
	mux := routeInit()

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
	mux := routeInit()

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

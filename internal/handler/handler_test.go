package handler

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/eshadow1/shortener/internal/configs"
	"github.com/eshadow1/shortener/internal/loggers"
	"github.com/eshadow1/shortener/internal/model"
	mockhandler "github.com/eshadow1/shortener/mocks/handler"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const (
	correctShort = "42b3e75f"
	correctURL   = "https://practicum.yandex.ru/"
)

func TestHandler_GetOrigin(t *testing.T) {
	cfg := &configs.Config{
		Addr:    configs.DefaultAddr,
		BaseURL: configs.DefaultBaseURL,
	}

	tests := []struct {
		name             string
		method           string
		url              string
		expectedStatus   int
		expectedLocation string
	}{
		{
			name:             "success",
			method:           http.MethodGet,
			url:              "/" + correctShort,
			expectedStatus:   http.StatusTemporaryRedirect,
			expectedLocation: correctURL,
		},
		{
			name:             "bad_method",
			method:           http.MethodPost,
			url:              "/" + correctShort,
			expectedStatus:   http.StatusMethodNotAllowed,
			expectedLocation: "",
		},
		{
			name:             "bad_url",
			method:           http.MethodGet,
			url:              "/abcd",
			expectedStatus:   http.StatusBadRequest,
			expectedLocation: "",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := httptest.NewRequestWithContext(t.Context(), test.method, test.url, http.NoBody)

			rsCtx := chi.NewRouteContext()
			rsCtx.URLParams.Add("shortURL", strings.TrimPrefix(test.url, "/"))

			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rsCtx))

			w := httptest.NewRecorder()

			mc := mockhandler.NewMockChecker(t)
			mc.On("ConnectDB", mock.Anything).Return(true).Maybe()

			ms := mockhandler.NewMockService(t)
			ms.On("GetOriginalURL", req.Context(), model.ShortenInfo{ShortURL: correctShort}).Return(model.OriginalInfo{OriginalURL: correctURL}, nil).Maybe()
			ms.On("GetOriginalURL", req.Context(), mock.Anything).Return(model.OriginalInfo{}, errors.New("short not found")).Maybe()
			h := NewHandler(cfg, ms, mc)

			h.GetOrigin(w, req)
			assert.Equal(t, test.expectedStatus, w.Code)
			assert.Equal(t, test.expectedLocation, w.Header().Get("Location"))
		})
	}
}

func TestHandler_PostCreate(t *testing.T) {
	cfg := &configs.Config{
		Addr:    configs.DefaultAddr,
		BaseURL: configs.DefaultBaseURL,
	}

	tests := []struct {
		name           string
		method         string
		body           string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "success",
			method:         http.MethodPost,
			body:           correctURL,
			expectedStatus: http.StatusCreated,
			expectedBody:   configs.DefaultBaseURL + "/" + correctShort,
		},
		{
			name:           "bad_method",
			method:         http.MethodGet,
			body:           correctURL,
			expectedStatus: http.StatusMethodNotAllowed,
			expectedBody:   http.StatusText(http.StatusBadRequest) + "\n",
		},
		{
			name:           "bad_body",
			method:         http.MethodPost,
			body:           "",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   http.StatusText(http.StatusBadRequest) + "\n",
		},
		{
			name:           "error_create_short_url",
			method:         http.MethodPost,
			body:           "practicum.yandex.ru",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   http.StatusText(http.StatusBadRequest) + "\n",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := httptest.NewRequestWithContext(t.Context(), test.method, "/", strings.NewReader(test.body))
			req.Header.Set("Content-Type", "text/plain")

			w := httptest.NewRecorder()

			mc := mockhandler.NewMockChecker(t)
			mc.On("ConnectDB", mock.Anything).Return(nil).Maybe()

			ms := mockhandler.NewMockService(t)
			ms.On("CreateShortURL", t.Context(), []model.OriginalInfo{{OriginalURL: correctURL}}).Return([]model.ShortenInfo{{ShortURL: correctShort}}, nil).Maybe()
			ms.On("CreateShortURL", t.Context(), mock.Anything).Return([]model.ShortenInfo{}, errors.New("bad request")).Maybe()
			h := NewHandler(cfg, ms, mc)

			h.PostCreate(w, req)

			body, err := io.ReadAll(w.Body)
			require.NoError(t, err)

			assert.Equal(t, test.expectedStatus, w.Code)
			assert.Equal(t, test.expectedBody, string(body))
		})
	}
}

func TestHandler_PostShorten(t *testing.T) {
	cfg := &configs.Config{
		Addr:    configs.DefaultAddr,
		BaseURL: configs.DefaultBaseURL,
	}

	tests := []struct {
		name                string
		method              string
		body                string
		headerContentType   string
		expectedStatus      int
		expectedContentType string
		expectedBody        string
	}{
		{
			name:                "success",
			method:              http.MethodPost,
			body:                fmt.Sprintf("{\"url\":%q}", correctURL),
			headerContentType:   "application/json",
			expectedStatus:      http.StatusCreated,
			expectedContentType: "application/json",
			expectedBody:        fmt.Sprintf("{\"result\":%q}", configs.DefaultBaseURL+"/"+correctShort),
		},
		{
			name:                "bad_method",
			method:              http.MethodGet,
			body:                fmt.Sprintf("{\"url\":%q}", correctURL),
			headerContentType:   "application/json",
			expectedStatus:      http.StatusMethodNotAllowed,
			expectedContentType: "text/plain; charset=utf-8",
			expectedBody:        http.StatusText(http.StatusBadRequest) + "\n",
		},
		{
			name:                "bad_body",
			method:              http.MethodPost,
			body:                "",
			headerContentType:   "application/json",
			expectedStatus:      http.StatusBadRequest,
			expectedContentType: "text/plain; charset=utf-8",
			expectedBody:        http.StatusText(http.StatusBadRequest) + "\n",
		},
		{
			name:                "error_create_short_url",
			method:              http.MethodPost,
			body:                "practicum.yandex.ru",
			headerContentType:   "application/json",
			expectedStatus:      http.StatusBadRequest,
			expectedContentType: "text/plain; charset=utf-8",
			expectedBody:        http.StatusText(http.StatusBadRequest) + "\n",
		},
		{
			name:                "error_content_type",
			method:              http.MethodPost,
			body:                "practicum.yandex.ru",
			headerContentType:   "application/text",
			expectedStatus:      http.StatusBadRequest,
			expectedContentType: "text/plain; charset=utf-8",
			expectedBody:        http.StatusText(http.StatusBadRequest) + "\n",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := httptest.NewRequestWithContext(t.Context(), test.method, "/api/shorten", strings.NewReader(test.body))
			req.Header.Set("Content-Type", test.headerContentType)

			w := httptest.NewRecorder()

			mc := mockhandler.NewMockChecker(t)
			mc.On("ConnectDB", mock.Anything).Return(nil).Maybe()

			ms := mockhandler.NewMockService(t)
			ms.On("CreateShortURL", t.Context(), []model.OriginalInfo{{OriginalURL: correctURL}}).Return([]model.ShortenInfo{{ShortURL: correctShort}}, nil).Maybe()
			ms.On("CreateShortURL", t.Context(), mock.Anything).Return(model.ShortenInfo{}, errors.New("bad request")).Maybe()
			h := NewHandler(cfg, ms, mc)

			h.PostShorten(w, req)

			body, err := io.ReadAll(w.Body)
			require.NoError(t, err)

			assert.Equal(t, test.expectedStatus, w.Code)
			assert.Equal(t, test.expectedContentType, w.Header().Get("Content-Type"))
			assert.Equal(t, test.expectedBody, string(body))
		})
	}
}

func TestHandler_PostShortenBatch(t *testing.T) {
	cfg := &configs.Config{
		Addr:    configs.DefaultAddr,
		BaseURL: configs.DefaultBaseURL,
	}

	tests := []struct {
		name                string
		method              string
		body                string
		headerContentType   string
		expectedStatus      int
		expectedContentType string
		expectedBody        string
	}{
		{
			name:                "success",
			method:              http.MethodPost,
			body:                fmt.Sprintf("[{\"original_url\":%q,\"correlation_id\":\"1\"}]", correctURL),
			headerContentType:   "application/json",
			expectedStatus:      http.StatusCreated,
			expectedContentType: "application/json",
			expectedBody:        fmt.Sprintf("[{\"short_url\":%q,\"correlation_id\":\"1\"}]", configs.DefaultBaseURL+"/"+correctShort),
		},
		{
			name:                "bad_method",
			method:              http.MethodGet,
			body:                fmt.Sprintf("{\"url\":%q}", correctURL),
			headerContentType:   "application/json",
			expectedStatus:      http.StatusMethodNotAllowed,
			expectedContentType: "text/plain; charset=utf-8",
			expectedBody:        http.StatusText(http.StatusBadRequest) + "\n",
		},
		{
			name:                "bad_body",
			method:              http.MethodPost,
			body:                "",
			headerContentType:   "application/json",
			expectedStatus:      http.StatusBadRequest,
			expectedContentType: "text/plain; charset=utf-8",
			expectedBody:        http.StatusText(http.StatusBadRequest) + "\n",
		},
		{
			name:                "error_create_short_url",
			method:              http.MethodPost,
			body:                "practicum.yandex.ru",
			headerContentType:   "application/json",
			expectedStatus:      http.StatusBadRequest,
			expectedContentType: "text/plain; charset=utf-8",
			expectedBody:        http.StatusText(http.StatusBadRequest) + "\n",
		},
		{
			name:                "error_content_type",
			method:              http.MethodPost,
			body:                "practicum.yandex.ru",
			headerContentType:   "application/text",
			expectedStatus:      http.StatusBadRequest,
			expectedContentType: "text/plain; charset=utf-8",
			expectedBody:        http.StatusText(http.StatusBadRequest) + "\n",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := httptest.NewRequestWithContext(t.Context(), test.method, "/api/shorten/batch", strings.NewReader(test.body))
			req.Header.Set("Content-Type", test.headerContentType)

			w := httptest.NewRecorder()

			mc := mockhandler.NewMockChecker(t)
			mc.On("ConnectDB", mock.Anything).Return(nil).Maybe()

			ms := mockhandler.NewMockService(t)
			ms.On("CreateShortURL", t.Context(), []model.OriginalInfo{{OriginalURL: correctURL, CorrelationID: "1"}}).Return([]model.ShortenInfo{{ShortURL: correctShort, CorrelationID: "1"}}, nil).Maybe()
			ms.On("CreateShortURL", t.Context(), mock.Anything).Return(model.ShortenInfo{}, errors.New("bad request")).Maybe()
			h := NewHandler(cfg, ms, mc)

			h.PostShortenBatch(w, req)

			body, err := io.ReadAll(w.Body)
			require.NoError(t, err)

			assert.Equal(t, test.expectedStatus, w.Code)
			assert.Equal(t, test.expectedContentType, w.Header().Get("Content-Type"))
			assert.Equal(t, test.expectedBody, string(body))
		})
	}
}

func TestHandler_GetCheckDB(t *testing.T) {
	cfg := &configs.Config{
		Addr:    configs.DefaultAddr,
		BaseURL: configs.DefaultBaseURL,
	}
	errLog := loggers.CreateLogger("Error")
	require.NoError(t, errLog)

	tests := []struct {
		name           string
		method         string
		resultCheck    error
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "success",
			method:         http.MethodGet,
			resultCheck:    nil,
			expectedStatus: http.StatusOK,
			expectedBody:   http.StatusText(http.StatusOK) + "\n",
		},
		{
			name:           "bad_method",
			method:         http.MethodPost,
			resultCheck:    fmt.Errorf("error"),
			expectedStatus: http.StatusMethodNotAllowed,
			expectedBody:   http.StatusText(http.StatusMethodNotAllowed) + "\n",
		},
		{
			name:           "bad_connection",
			method:         http.MethodGet,
			resultCheck:    fmt.Errorf("error"),
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   http.StatusText(http.StatusInternalServerError) + "\n",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := httptest.NewRequestWithContext(t.Context(), test.method, "/ping", http.NoBody)
			req.Header.Set("Content-Type", "text/plain")

			w := httptest.NewRecorder()

			mc := mockhandler.NewMockChecker(t)
			mc.On("ConnectDB", mock.Anything).Return(test.resultCheck).Maybe()

			ms := mockhandler.NewMockService(t)
			ms.On("CreateShortURL", t.Context(), mock.Anything).Return(model.ShortenInfo{}, nil).Maybe()
			h := NewHandler(cfg, ms, mc)

			h.GetCheckDB(w, req)

			body, err := io.ReadAll(w.Body)
			require.NoError(t, err)

			assert.Equal(t, test.expectedStatus, w.Code)
			assert.Equal(t, test.expectedBody, string(body))
		})
	}
}

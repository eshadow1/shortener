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
	"github.com/eshadow1/shortener/internal/model"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const (
	correctShort = "42b3e75f"
	correctURL   = "https://practicum.yandex.ru/"
)

type MockService struct {
	mock.Mock
}

func (m *MockService) GetOriginalURL(ctx context.Context, short model.ShortenInfo) (model.OriginalInfo, error) {
	args := m.Called(ctx, short)
	return args.Get(0).(model.OriginalInfo), args.Error(1)
}

func (m *MockService) CreateShortUrl(ctx context.Context, origin model.OriginalInfo) (model.ShortenInfo, error) {
	args := m.Called(ctx, origin)
	return args.Get(0).(model.ShortenInfo), args.Error(1)
}

func TestHandler_GetOrigin(t *testing.T) {
	cfg := &configs.Config{
		Addr:    configs.DefaultAddr,
		BaseUrl: configs.DefaultBaseUrl,
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

			ms := new(MockService)
			ms.On("GetOriginalURL", req.Context(), model.ShortenInfo{ShortURL: correctShort}).Return(model.OriginalInfo{OriginalURL: correctURL}, nil)
			ms.On("GetOriginalURL", req.Context(), mock.Anything).Return(model.OriginalInfo{}, errors.New("short not found"))
			h := NewHandler(cfg, ms)

			h.GetOrigin(w, req)
			assert.Equal(t, test.expectedStatus, w.Code)
			assert.Equal(t, test.expectedLocation, w.Header().Get("Location"))
		})
	}
}

func TestHandler_PostCreate(t *testing.T) {
	cfg := &configs.Config{
		Addr:    configs.DefaultAddr,
		BaseUrl: configs.DefaultBaseUrl,
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
			expectedBody:   configs.DefaultBaseUrl + "/" + correctShort,
		},
		{
			name:           "bad_method",
			method:         http.MethodGet,
			body:           correctURL,
			expectedStatus: http.StatusMethodNotAllowed,
			expectedBody:   "Bad request\n",
		},
		{
			name:           "bad_body",
			method:         http.MethodPost,
			body:           "",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Bad request\n",
		},
		{
			name:           "error_create_short_url",
			method:         http.MethodPost,
			body:           "practicum.yandex.ru",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Bad request\n",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := httptest.NewRequestWithContext(t.Context(), test.method, "/", strings.NewReader(test.body))
			req.Header.Set("Content-Type", "text/plain")

			w := httptest.NewRecorder()

			ms := new(MockService)
			ms.On("CreateShortUrl", t.Context(), model.OriginalInfo{OriginalURL: correctURL}).Return(model.ShortenInfo{ShortURL: correctShort}, nil)
			ms.On("CreateShortUrl", t.Context(), mock.Anything).Return(model.ShortenInfo{}, errors.New("bad request"))
			h := NewHandler(cfg, ms)

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
		BaseUrl: configs.DefaultBaseUrl,
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
			expectedBody:        fmt.Sprintf("{\"result\":%q}", configs.DefaultBaseUrl+"/"+correctShort),
		},
		{
			name:                "bad_method",
			method:              http.MethodGet,
			body:                fmt.Sprintf("{\"url\":%q}", correctURL),
			headerContentType:   "application/json",
			expectedStatus:      http.StatusMethodNotAllowed,
			expectedContentType: "text/plain; charset=utf-8",
			expectedBody:        "Bad request\n",
		},
		{
			name:                "bad_body",
			method:              http.MethodPost,
			body:                "",
			headerContentType:   "application/json",
			expectedStatus:      http.StatusBadRequest,
			expectedContentType: "text/plain; charset=utf-8",
			expectedBody:        "Bad request\n",
		},
		{
			name:                "error_create_short_url",
			method:              http.MethodPost,
			body:                "practicum.yandex.ru",
			headerContentType:   "application/json",
			expectedStatus:      http.StatusBadRequest,
			expectedContentType: "text/plain; charset=utf-8",
			expectedBody:        "Bad request\n",
		},
		{
			name:                "error_content_type",
			method:              http.MethodPost,
			body:                "practicum.yandex.ru",
			headerContentType:   "application/text",
			expectedStatus:      http.StatusBadRequest,
			expectedContentType: "text/plain; charset=utf-8",
			expectedBody:        "Bad request\n",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := httptest.NewRequestWithContext(t.Context(), test.method, "/api/shorten", strings.NewReader(test.body))
			req.Header.Set("Content-Type", test.headerContentType)

			w := httptest.NewRecorder()

			ms := new(MockService)
			ms.On("CreateShortUrl", t.Context(), model.OriginalInfo{OriginalURL: correctURL}).Return(model.ShortenInfo{ShortURL: correctShort}, nil)
			ms.On("CreateShortUrl", t.Context(), mock.Anything).Return(model.ShortenInfo{}, errors.New("bad request"))
			h := NewHandler(cfg, ms)

			h.PostShorten(w, req)

			body, err := io.ReadAll(w.Body)
			require.NoError(t, err)

			assert.Equal(t, test.expectedStatus, w.Code)
			assert.Equal(t, test.expectedContentType, w.Header().Get("Content-Type"))
			assert.Equal(t, test.expectedBody, string(body))
		})
	}
}

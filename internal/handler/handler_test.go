package handler

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/eshadow1/shortener/internal/configs"
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

func (m *MockService) GetOriginalURL(short string) (string, error) {
	args := m.Called(short)
	return args.String(0), args.Error(1)
}

func (m *MockService) CreateShortUrl(origin string) (string, error) {
	args := m.Called(origin)
	return args.String(0), args.Error(1)
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
			expectedStatus:   http.StatusBadRequest,
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
			ms.On("GetOriginalURL", correctShort).Return(correctURL, nil)
			ms.On("GetOriginalURL", mock.Anything).Return("", errors.New("short not found"))
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
			expectedBody:   configs.DefaultBaseUrl + correctShort,
		},
		{
			name:           "bad_method",
			method:         http.MethodGet,
			body:           correctURL,
			expectedStatus: http.StatusBadRequest,
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
			ms.On("CreateShortUrl", correctURL).Return(correctShort, nil)
			ms.On("CreateShortUrl", mock.Anything).Return("", errors.New("bad request"))
			h := NewHandler(cfg, ms)

			h.PostCreate(w, req)

			body, err := io.ReadAll(w.Body)
			require.NoError(t, err)

			assert.Equal(t, test.expectedStatus, w.Code)
			assert.Equal(t, test.expectedBody, string(body))
		})
	}
}

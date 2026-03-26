package handler

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
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
			url:              "/abcdefgh",
			expectedStatus:   http.StatusTemporaryRedirect,
			expectedLocation: "https://practicum.yandex.ru/",
		},
		{
			name:             "bad_method",
			method:           http.MethodPost,
			url:              "/abcdefgh",
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
			w := httptest.NewRecorder()

			ms := new(MockService)
			ms.On("GetOriginalURL", "abcdefgh").Return("https://practicum.yandex.ru/", nil)
			ms.On("GetOriginalURL", mock.Anything).Return("", errors.New("short not found"))
			h := NewHandler(ms)

			h.GetOrigin(w, req)
			assert.Equal(t, test.expectedStatus, w.Code)
			assert.Equal(t, test.expectedLocation, w.Header().Get("Location"))
		})
	}
}

func TestHandler_PostCreate(t *testing.T) {
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
			body:           "https://practicum.yandex.ru/",
			expectedStatus: http.StatusCreated,
			expectedBody:   "http://localhost:8080/abcdefgh",
		},
		{
			name:           "bad_method",
			method:         http.MethodGet,
			body:           "https://practicum.yandex.ru/",
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
			ms.On("CreateShortUrl", "https://practicum.yandex.ru/").Return("abcdefgh", nil)
			ms.On("CreateShortUrl", mock.Anything).Return("", errors.New("bad request"))
			h := NewHandler(ms)

			h.PostCreate(w, req)

			body, err := io.ReadAll(w.Body)
			require.NoError(t, err)

			assert.Equal(t, test.expectedStatus, w.Code)
			assert.Equal(t, test.expectedBody, string(body))
		})
	}
}

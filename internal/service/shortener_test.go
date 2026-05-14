package service

import (
	"errors"
	"testing"
	"time"

	"github.com/eshadow1/shortener/internal/configs"
	"github.com/eshadow1/shortener/internal/model"
	mockservice "github.com/eshadow1/shortener/mocks/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const (
	correctShort   = "42b3e75f"
	correctURL     = "https://practicum.yandex.ru/"
	testBufferSize = 10
	testBatchSize  = 10
	testTimeout    = 10 * time.Second
)

func TestShortenerService_CreateShortUrl(t *testing.T) {
	cfg := configs.ServiceConfig{
		BufferSizeChan: testBufferSize,
		BatchSize:      testBatchSize,
		FlushInterval:  testTimeout,
	}
	tests := []struct {
		name          string
		url           []model.OriginalInfo
		expectedShort []model.ShortenInfo
		expectedError error
	}{
		{
			name:          "success_create",
			url:           []model.OriginalInfo{{OriginalURL: correctURL}},
			expectedShort: []model.ShortenInfo{{ShortURL: correctShort}},
			expectedError: nil,
		},
		{
			name:          "error_create",
			url:           []model.OriginalInfo{{OriginalURL: "https"}},
			expectedShort: []model.ShortenInfo{{ShortURL: "3e194352"}},
			expectedError: errors.New("don't save"),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mr := mockservice.NewMockRepository(t)
			mr.On("Save", t.Context(), []model.URLInfo{{ShortURL: correctShort, OriginalURL: correctURL}}).Return(nil).Maybe()
			mr.On("Save", t.Context(), mock.Anything, mock.Anything).Return(errors.New("don't save")).Maybe()
			s := NewShortenerService(mr, cfg)

			short, errSave := s.CreateShortURL(t.Context(), test.url)
			if test.expectedError != nil {
				assert.Equal(t, test.expectedError, errSave)
			} else {
				require.NoError(t, errSave)
			}
			assert.Equal(t, test.expectedShort, short)
		})
	}
}

func TestShortenerService_GetShortUrl(t *testing.T) {
	cfg := configs.ServiceConfig{
		BufferSizeChan: testBufferSize,
		BatchSize:      testBatchSize,
		FlushInterval:  testTimeout,
	}
	tests := []struct {
		name             string
		short            model.ShortenInfo
		expectedOriginal model.OriginalInfo
		expectedError    error
	}{
		{
			name:             "success",
			short:            model.ShortenInfo{ShortURL: correctShort},
			expectedOriginal: model.OriginalInfo{OriginalURL: correctURL},
			expectedError:    nil,
		},
		{
			name:             "error_get",
			short:            model.ShortenInfo{ShortURL: "5e056c50"},
			expectedOriginal: model.OriginalInfo{OriginalURL: ""},
			expectedError:    errors.New("not found"),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mr := mockservice.NewMockRepository(t)
			mr.On("Get", t.Context(), correctShort).Return(model.UserURL{OriginalURL: correctURL, ShortURL: correctShort}, nil).Maybe()
			mr.On("Get", t.Context(), mock.Anything).Return(model.UserURL{}, errors.New("not found")).Maybe()
			s := NewShortenerService(mr, cfg)

			original, errSave := s.GetOriginalURL(t.Context(), test.short)
			if test.expectedError != nil {
				assert.Equal(t, test.expectedError, errSave)
			} else {
				require.NoError(t, errSave)
			}
			assert.Equal(t, test.expectedOriginal, original)
		})
	}
}

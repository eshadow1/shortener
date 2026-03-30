package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const (
	correctShort = "42b3e75f"
	correctURL   = "https://practicum.yandex.ru/"
)

type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) Save(ctx context.Context, key, value string) error {
	args := m.Called(ctx, key, value)
	return args.Error(0)
}

func (m *MockRepository) Get(ctx context.Context, key string) (string, error) {
	args := m.Called(ctx, key)
	return args.String(0), args.Error(1)
}

func TestShortenerService_CreateShortUrl(t *testing.T) {
	tests := []struct {
		name          string
		url           string
		expectedShort string
		expectedError error
	}{
		{
			name:          "success_create",
			url:           correctURL,
			expectedShort: correctShort,
			expectedError: nil,
		},
		{
			name:          "error_create",
			url:           "https",
			expectedShort: "3e194352",
			expectedError: errors.New("don't save"),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mr := new(MockRepository)
			mr.On("Save", t.Context(), correctShort, correctURL).Return(nil)
			mr.On("Save", t.Context(), mock.Anything, mock.Anything).Return(errors.New("don't save"))
			s := NewShortenerService(mr)

			short, errSave := s.CreateShortUrl(t.Context(), test.url)
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
	tests := []struct {
		name             string
		short            string
		expectedOriginal string
		expectedError    error
	}{
		{
			name:             "success",
			short:            correctShort,
			expectedOriginal: correctURL,
			expectedError:    nil,
		},
		{
			name:             "error_get",
			short:            "5e056c50",
			expectedOriginal: "",
			expectedError:    errors.New("not found"),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mr := new(MockRepository)
			mr.On("Get", t.Context(), correctShort).Return(correctURL, nil)
			mr.On("Get", t.Context(), mock.Anything).Return("", errors.New("not found"))
			s := NewShortenerService(mr)

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

package service

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) Save(key, value string) error {
	args := m.Called(key, value)
	return args.Error(0)
}

func (m *MockRepository) Get(key string) (string, error) {
	args := m.Called(key)
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
			url:           "https://practicum.yandex.ru/",
			expectedShort: "0dd19817",
			expectedError: nil,
		},
		{
			name:          "error_create",
			url:           "https",
			expectedShort: "5e056c50",
			expectedError: errors.New("don't save"),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mr := new(MockRepository)
			mr.On("Save", "0dd19817", "https://practicum.yandex.ru/").Return(nil)
			mr.On("Save", mock.Anything, mock.Anything).Return(errors.New("don't save"))
			s := NewShortenerService(mr)

			short, errSave := s.CreateShortUrl(test.url)
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
			short:            "0dd19817",
			expectedOriginal: "https://practicum.yandex.ru/",
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
			mr.On("Get", "0dd19817").Return("https://practicum.yandex.ru/", nil)
			mr.On("Get", mock.Anything).Return("", errors.New("not found"))
			s := NewShortenerService(mr)

			original, errSave := s.GetOriginalURL(test.short)
			if test.expectedError != nil {
				assert.Equal(t, test.expectedError, errSave)
			} else {
				require.NoError(t, errSave)
			}
			assert.Equal(t, test.expectedOriginal, original)
		})
	}
}

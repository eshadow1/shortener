package repository

import (
	"context"
	"errors"
	"testing"

	"github.com/eshadow1/shortener/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	defaultOriginal    = "https://practicum.yandex.ru/"
	defaultShort       = "abcdefgh"
	defaultStoragePath = ""
	defaultUUID        = "1234-test-uuid"
)

func TestRepository_Get(t *testing.T) {
	m := NewMemoryRepository(defaultStoragePath)
	ctx := context.WithValue(t.Context(), model.UserIDContextKey, defaultUUID)
	errSave := m.Save(ctx, []model.URLInfo{{ShortURL: defaultShort, OriginalURL: defaultOriginal}})
	require.NoError(t, errSave)

	tests := []struct {
		name             string
		short            string
		expectedOriginal model.UserURL
		expectedError    error
	}{
		{

			name:  "success",
			short: defaultShort,
			expectedOriginal: model.UserURL{
				OriginalURL: defaultOriginal,
				ShortURL:    defaultShort,
				IsDeleted:   false,
			},
			expectedError: nil,
		},
		{
			name:  "error_get",
			short: "not_found",
			expectedOriginal: model.UserURL{
				OriginalURL: "",
				ShortURL:    "",
				IsDeleted:   false,
			},
			expectedError: errors.New("short not found"),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			original, errGet := m.Get(ctx, test.short)
			if test.expectedError != nil {
				assert.Equal(t, test.expectedError, errGet)
			} else {
				require.NoError(t, errGet)
			}
			assert.Equal(t, test.expectedOriginal, original)
		})
	}
}

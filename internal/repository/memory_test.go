package repository

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"github.com/eshadow1/shortener/internal/loggers"
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

func TestMemoryRepository_GetUserURLs(t *testing.T) {
	m := NewMemoryRepository(defaultStoragePath)
	ctx := context.WithValue(t.Context(), model.UserIDContextKey, defaultUUID)
	errSave := m.Save(ctx, []model.URLInfo{{ShortURL: defaultShort, OriginalURL: defaultOriginal}})
	require.NoError(t, errSave)

	tests := []struct {
		name             string
		short            string
		userID           string
		expectedOriginal []model.UserURL
	}{
		{
			name:   "success",
			short:  defaultShort,
			userID: defaultUUID,
			expectedOriginal: []model.UserURL{
				{
					OriginalURL: defaultOriginal,
					ShortURL:    defaultShort,
					IsDeleted:   false,
				},
			},
		},
		{
			name:             "error_get",
			short:            "not_found",
			userID:           "",
			expectedOriginal: []model.UserURL{},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctxT := context.WithValue(t.Context(), model.UserIDContextKey, test.userID)
			original, errGet := m.GetUserURLs(ctxT)
			require.NoError(t, errGet)
			assert.Equal(t, test.expectedOriginal, original)
		})
	}
}

func TestMemoryRepository_DeleteUserURLs(t *testing.T) {
	m := NewMemoryRepository(defaultStoragePath)
	ctx := context.WithValue(t.Context(), model.UserIDContextKey, defaultUUID)
	errSave := m.Save(ctx, []model.URLInfo{{ShortURL: defaultShort, OriginalURL: defaultOriginal}})
	require.NoError(t, errSave)

	tests := []struct {
		name   string
		short  []string
		userID string
	}{
		{
			name:   "success",
			short:  []string{defaultShort},
			userID: defaultUUID,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			errGet := m.DeleteUserURLs(ctx, test.userID, test.short)
			require.NoError(t, errGet)
		})
	}
}

func TestMemoryRepository_SaveUserURLs(t *testing.T) {
	errLog := loggers.CreateLogger("error")
	require.NoError(t, errLog)

	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "data.txt")
	memBeforeSave := NewMemoryRepository(filePath)
	ctx := context.WithValue(t.Context(), model.UserIDContextKey, defaultUUID)

	errSave := memBeforeSave.Save(ctx, []model.URLInfo{{ShortURL: defaultShort, OriginalURL: defaultOriginal}})
	require.NoError(t, errSave)

	memBeforeSave.Close()

	memAfterSave := NewMemoryRepository(filePath)
	origin, errGet := memAfterSave.Get(ctx, defaultShort)
	require.NoError(t, errGet)
	assert.Equal(t, defaultShort, origin.ShortURL)
	assert.Equal(t, defaultOriginal, origin.OriginalURL)

	memAfterSave.Close()
}

package repository

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	defaultOriginal = "https://practicum.yandex.ru/"
	defaultShort    = "abcdefgh"
)

func TestRepository_Get(t *testing.T) {
	m := NewMemoryRepository()
	errSave := m.Save(defaultShort, defaultOriginal)
	require.NoError(t, errSave)

	tests := []struct {
		name             string
		short            string
		expectedOriginal string
		expectedError    error
	}{
		{

			name:             "success",
			short:            defaultShort,
			expectedOriginal: defaultOriginal,
			expectedError:    nil,
		},
		{
			name:             "error_get",
			short:            "not_found",
			expectedOriginal: "",
			expectedError:    errors.New("short not found"),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			original, errGet := m.Get(test.short)
			if test.expectedError != nil {
				assert.Equal(t, test.expectedError, errGet)
			} else {
				require.NoError(t, errGet)
			}
			assert.Equal(t, test.expectedOriginal, original)
		})
	}
}

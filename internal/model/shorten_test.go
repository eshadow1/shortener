package model

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOriginalInfo_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name             string
		originalData     []byte
		expectedOriginal OriginalInfo
		expectedError    string
	}{
		{
			name:             "success_url",
			originalData:     []byte("{\"url\":\"test\"}"),
			expectedOriginal: OriginalInfo{OriginalURL: "test"},
			expectedError:    "",
		},
		{
			name:             "success_full_url",
			originalData:     []byte("{\"url\":\"test\",\"correlation_id\":\"1\"}"),
			expectedOriginal: OriginalInfo{OriginalURL: "test", CorrelationID: "1"},
			expectedError:    "",
		},
		{
			name:             "success_original_url",
			originalData:     []byte("{\"original_url\":\"test\"}"),
			expectedOriginal: OriginalInfo{OriginalURL: "test"},
			expectedError:    "",
		},
		{
			name:             "success_full_result",
			originalData:     []byte("{\"original_url\":\"test\",\"correlation_id\":\"1\"}"),
			expectedOriginal: OriginalInfo{OriginalURL: "test", CorrelationID: "1"},
			expectedError:    "",
		},
		{
			name:             "incorrect_data",
			originalData:     []byte("}"),
			expectedOriginal: OriginalInfo{},
			expectedError:    "invalid character '}' looking for beginning of value",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var tempOriginal OriginalInfo
			err := json.Unmarshal(test.originalData, &tempOriginal)
			if test.expectedError != "" {
				assert.Equal(t, test.expectedError, err.Error())
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, test.expectedOriginal, tempOriginal)
		})
	}
}

func TestShortenInfo_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name            string
		shortData       []byte
		expectedShorten ShortenInfo
		expectedError   string
	}{
		{
			name:            "success_url",
			shortData:       []byte("{\"short_url\":\"test\"}"),
			expectedShorten: ShortenInfo{ShortURL: "test"},
			expectedError:   "",
		},
		{
			name:            "success_full_url",
			shortData:       []byte("{\"short_url\":\"test\",\"correlation_id\":\"1\"}"),
			expectedShorten: ShortenInfo{ShortURL: "test", CorrelationID: "1"},
			expectedError:   "",
		},
		{
			name:            "success_result",
			shortData:       []byte("{\"result\":\"test\"}"),
			expectedShorten: ShortenInfo{ShortURL: "test"},
			expectedError:   "",
		},
		{
			name:            "success_full_result",
			shortData:       []byte("{\"result\":\"test\",\"correlation_id\":\"1\"}"),
			expectedShorten: ShortenInfo{ShortURL: "test", CorrelationID: "1"},
			expectedError:   "",
		},
		{
			name:            "incorrect_data",
			shortData:       []byte("}"),
			expectedShorten: ShortenInfo{},
			expectedError:   "invalid character '}' looking for beginning of value",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var tempShorten ShortenInfo
			err := json.Unmarshal(test.shortData, &tempShorten)
			if test.expectedError != "" {
				assert.Equal(t, test.expectedError, err.Error())
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, test.expectedShorten, tempShorten)
		})
	}
}

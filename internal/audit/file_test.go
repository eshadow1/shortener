package audit

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/eshadow1/shortener/internal/loggers"
	"github.com/eshadow1/shortener/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewFileObserver проверяет создание fileObserver.
func TestFileObserver_NewFileObserver(t *testing.T) {
	tmpDir := t.TempDir()
	errLog := loggers.CreateLogger("error")
	require.NoError(t, errLog)

	tests := []struct {
		name     string
		filePath string
		wantNil  bool
	}{
		{
			name:     "empty path",
			filePath: "",
			wantNil:  true,
		},
		{
			name:     "valid path",
			filePath: filepath.Join(tmpDir, "audit.log"),
			wantNil:  false,
		},
		{
			name:     "invalid directory",
			filePath: "/nonexistent/dir/audit.log",
			wantNil:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obs := NewFileObserver(tt.filePath)
			if tt.wantNil {
				assert.Nil(t, obs)
			} else {
				require.NotNil(t, obs)
				assert.NotNil(t, obs.file)
				obs.Close()
			}
		})
	}
}

// TestNotify проверяет запись событий в файл.
func TestFileObserver_Notify(t *testing.T) {
	errLog := loggers.CreateLogger("error")
	require.NoError(t, errLog)

	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "audit.log")

	obs := NewFileObserver(filePath)
	require.NotNil(t, obs)
	defer obs.Close()

	tests := []struct {
		name  string
		event model.Event
	}{
		{
			name: "simple event",
			event: model.Event{
				Action: model.Follow,
				URL:    "test",
			},
		},
		{
			name: "event with nested data",
			event: model.Event{
				Action: model.Follow,
				URL:    "test2",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obs.Notify(tt.event)

			data, err := os.ReadFile(filePath)
			require.NoError(t, err)

			lines := bytes.Split(bytes.TrimSpace(data), []byte{'\n'})
			require.NotEmpty(t, lines, "no lines in file")
			lastLine := lines[len(lines)-1]

			var got model.Event
			err = json.Unmarshal(lastLine, &got)
			require.NoError(t, err, "failed to unmarshal event")

			assert.Equal(t, tt.event, got, "event mismatch")
		})
	}
}

// TestClose проверяет закрытие файлового дескриптора.
func TestFileObserver_Close(t *testing.T) {
	errLog := loggers.CreateLogger("error")
	require.NoError(t, errLog)

	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "audit.log")

	obs := NewFileObserver(filePath)
	require.NotNil(t, obs, "failed to create observer")

	obs.Notify(model.Event{Action: model.Follow})
	obs.Close()
	obs.Notify(model.Event{Action: model.Shorten})

	finalData, err := os.ReadFile(filePath)
	require.NoError(t, err)

	lines := bytes.Split(bytes.TrimSpace(finalData), []byte{'\n'})
	assert.Len(t, lines, 1)
}

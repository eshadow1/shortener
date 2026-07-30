package audit

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/eshadow1/shortener/internal/loggers"
	"github.com/eshadow1/shortener/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewRemoteObserver проверяет создание remoteObserver.
func TestRemoteObserver_NewRemoteObserver(t *testing.T) {
	errLog := loggers.CreateLogger("error")
	require.NoError(t, errLog)

	tests := []struct {
		name    string
		url     string
		wantNil bool
	}{
		{
			name:    "empty URL",
			url:     "",
			wantNil: true,
		},
		{
			name:    "valid URL",
			url:     "http://example.com",
			wantNil: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obs := NewRemoteObserver(tt.url)
			if tt.wantNil {
				assert.Nil(t, obs)
			} else {
				require.NotNil(t, obs)
				assert.Equal(t, tt.url, obs.url)
				assert.NotNil(t, obs.client)
				assert.Equal(t, defaultRetryMax, obs.client.RetryMax)
				assert.Equal(t, defaultRetryWaitMin, obs.client.RetryWaitMin)
				assert.Equal(t, defaultRetryWaitMax, obs.client.RetryWaitMax)
				assert.NotNil(t, obs.client.CheckRetry)
				assert.NotNil(t, obs.client.HTTPClient)
				assert.Equal(t, defaultTimeout, obs.client.HTTPClient.Timeout)
			}
		})
	}
}

// TestNotify проверяет отправку событий на удалённый сервер.
func TestRemoteObserver_Notify(t *testing.T) {
	errLog := loggers.CreateLogger("error")
	require.NoError(t, errLog)

	var (
		mu             sync.Mutex
		receivedBody   []byte
		receivedHeader http.Header
		statusCode     int
	)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		receivedHeader = r.Header.Clone()
		body, _ := io.ReadAll(r.Body)
		receivedBody = body
		w.WriteHeader(statusCode)
	}))
	defer server.Close()

	obs := NewRemoteObserver(server.URL)
	require.NotNil(t, obs)
	defer obs.Close()

	tests := []struct {
		name       string
		event      model.Event
		statusCode int
	}{
		{
			name: "successful request",
			event: model.Event{
				Action: model.Follow,
				URL:    "test",
			},
			statusCode: http.StatusOK,
		},
		{
			name: "bad request",
			event: model.Event{
				Action: model.Follow,
				URL:    "test",
			},
			statusCode: http.StatusBadRequest,
		},
		{
			name: "internal server error",
			event: model.Event{
				Action: model.Shorten,
				URL:    "test2",
			},
			statusCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mu.Lock()
			receivedBody = nil
			receivedHeader = nil
			statusCode = tt.statusCode
			mu.Unlock()

			obs.Notify(tt.event)

			mu.Lock()
			defer mu.Unlock()
			require.NotNil(t, receivedBody)

			assert.Equal(t, "application/json", receivedHeader.Get("Content-Type"))

			var event model.Event
			err := json.Unmarshal(receivedBody, &event)
			require.NoError(t, err)
			assert.Equal(t, tt.event, event)
		})
	}
}

// TestClose проверяет, что Close не паникует.
func TestRemoteObserver_Close(t *testing.T) {
	obs := NewRemoteObserver("http://example.com")
	require.NotNil(t, obs)

	assert.NotPanics(t, func() { obs.Close() })
	assert.NotPanics(t, func() { obs.Close() })
}

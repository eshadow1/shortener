package audit

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/eshadow1/shortener/internal/loggers"
	"github.com/eshadow1/shortener/internal/model"

	"github.com/hashicorp/go-retryablehttp"
)

const (
	defaultTimeout      = 5 * time.Second
	defaultRetryMax     = 5
	defaultRetryWaitMin = 200 * time.Millisecond
	defaultRetryWaitMax = 25 * time.Second
)

// remoteObserver — наблюдатель, отправляющий события на удалённый сервер
type remoteObserver struct {
	url    string
	client *retryablehttp.Client
}

// NewRemoteObserver создает и возвращает нового удалённого наблюдателя,
// который отправляет события аудита по указанному URL методом POST.
func NewRemoteObserver(url string) *remoteObserver {
	if url == "" {
		loggers.Log.Info("No URL provided")
		return nil
	}

	retryClient := retryablehttp.NewClient()

	retryClient.RetryMax = defaultRetryMax
	retryClient.RetryWaitMin = defaultRetryWaitMin
	retryClient.RetryWaitMax = defaultRetryWaitMax
	retryClient.CheckRetry = retryPolicy
	retryClient.HTTPClient = &http.Client{
		Timeout: defaultTimeout,
	}

	loggers.Log.Info("Initializing remote observer", url)
	return &remoteObserver{
		url:    url,
		client: retryClient,
	}
}

// Notify отправляет переданное событие на удалённый сервер
// в формате JSON методом POST с заголовком Content-Type: application/json.
func (r *remoteObserver) Notify(event model.Event) {
	data, err := json.Marshal(event)
	if err != nil {
		loggers.Log.Error("Error serializing event", event, "error", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	req, err := retryablehttp.NewRequestWithContext(ctx, http.MethodPost, r.url, bytes.NewReader(data))
	if err != nil {
		loggers.Log.Error("Error create request event", event, "error", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := r.client.Do(req)
	if err != nil {
		loggers.Log.Error("Error posting event", event, "error", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		loggers.Log.Error("Error posting event", event, "status", resp.StatusCode)
	}
}

// Close закрывает открытый клиент.
func (r *remoteObserver) Close() {
	r.client.HTTPClient.CloseIdleConnections()
}

func retryPolicy(_ context.Context, resp *http.Response, err error) (bool, error) {
	if err != nil ||
		resp.StatusCode == http.StatusTooManyRequests ||
		resp.StatusCode >= http.StatusInternalServerError {
		return true, nil
	}

	return false, nil
}

package audit

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/eshadow1/shortener/internal/loggers"
	"github.com/eshadow1/shortener/internal/model"
)

const (
	defaultTimeout = 10 * time.Second
)

// remoteObserver — наблюдатель, отправляющий события на удалённый сервер
type remoteObserver struct {
	url    string
	client *http.Client
}

// NewRemoteObserver создает и возвращает нового удалённого наблюдателя,
// который отправляет события аудита по указанному URL методом POST.
func NewRemoteObserver(url string) *remoteObserver {
	if url == "" {
		loggers.Log.Info("No URL provided")
		return nil
	}
	loggers.Log.Info("Initializing remote observer", url)
	return &remoteObserver{
		url:    url,
		client: &http.Client{Timeout: defaultTimeout},
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

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, r.url, bytes.NewReader(data))
	if err != nil {
		log.Printf("[RemoteObserver] build request: %v", err)
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

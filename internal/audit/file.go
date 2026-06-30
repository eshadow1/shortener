package audit

import (
	"encoding/json"
	"os"
	"sync"

	"github.com/eshadow1/shortener/internal/loggers"
	"github.com/eshadow1/shortener/internal/model"
)

const (
	auditFilePerm os.FileMode = 0o644
)

// fileObserver — наблюдатель, записывающий события в файл
type fileObserver struct {
	filePath string
	mu       sync.Mutex
}

// NewFileObserver создает и возвращает новый файловый наблюдатель,
// который записывает события аудита в файл по указанному пути.
func NewFileObserver(filePath string) *fileObserver {
	if filePath == "" {
		loggers.Log.Info("No audit file path provided")
		return nil
	}
	loggers.Log.Info("Initializing remote observer: ", filePath)
	return &fileObserver{filePath: filePath}
}

// Notify записывает переданное событие аудита в файл в формате JSON,
// разделяя записи символом новой строки.
func (f *fileObserver) Notify(event model.Event) {
	f.mu.Lock()
	defer f.mu.Unlock()

	file, err := os.OpenFile(f.filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, auditFilePerm)
	if err != nil {
		loggers.Log.Error("File observer notify", "error open file", err.Error())
		return
	}
	defer file.Close()

	data, err := json.Marshal(event)
	if err != nil {
		loggers.Log.Error("Error serializing event", event, "error", err)
		return
	}
	data = append(data, '\n')

	if _, err = file.Write(data); err != nil {
		loggers.Log.Error("Error write file event", event, "error", err)
	}
}

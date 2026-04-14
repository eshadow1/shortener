package repository

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"sync"

	"github.com/eshadow1/shortener/internal/loggers"
	"github.com/eshadow1/shortener/internal/model"
)

var ErrShortNotFound = errors.New("short not found")

type memoryRepository struct {
	matchPairs  map[string]string
	storagePath string
	mu          sync.RWMutex
}

func NewMemoryRepository(storagePath string) *memoryRepository {
	return &memoryRepository{
		matchPairs:  loadData(storagePath),
		storagePath: storagePath,
		mu:          sync.RWMutex{},
	}
}

func (m *memoryRepository) Save(_ context.Context, short, origin string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.matchPairs[short]; !ok {
		m.matchPairs[short] = origin
		saveData(m.storagePath, m.matchPairs)
	}

	return nil
}

func (m *memoryRepository) Get(_ context.Context, short string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if val, ok := m.matchPairs[short]; ok {
		return val, nil
	}

	return "", ErrShortNotFound
}

func (m *memoryRepository) Close() {
	saveData(m.storagePath, m.matchPairs)
}

func saveData(storagePath string, data map[string]string) {
	if len(data) == 0 {
		return
	}
	temp := make([]model.FileStorage, 0, len(data))
	index := 1
	for short, origin := range data {
		temp = append(temp, model.FileStorage{
			UUID:     index,
			Short:    short,
			Original: origin,
		})
		index++
	}
	if storagePath == "" {
		return
	}

	tempString, errMarshal := json.Marshal(temp)
	if errMarshal != nil {
		loggers.Log.Errorf("Error marshal json: %s", errMarshal)
		return
	}

	f, errCreate := os.Create(storagePath)
	if errCreate != nil {
		loggers.Log.Errorf("Error create file %s: %s", storagePath, errCreate)
		return
	}
	defer f.Close()

	_, errWrite := f.Write(tempString)
	if errWrite != nil {
		loggers.Log.Errorf("Error write file %s: %s", storagePath, errWrite)
	}
}

func loadData(storagePath string) map[string]string {
	matchPairs := make(map[string]string)
	if storagePath == "" {
		return matchPairs
	}

	data, errRead := os.ReadFile(storagePath)
	if errRead != nil {
		loggers.Log.Errorf("Error reading file %s: %s", storagePath, errRead)
		return matchPairs
	}

	var temp []model.FileStorage
	errUnmarshal := json.Unmarshal(data, &temp)
	if errUnmarshal != nil {
		loggers.Log.Errorf("Error unmarshalling file %s: %s", storagePath, errUnmarshal)
		return matchPairs
	}

	for _, s := range temp {
		matchPairs[s.Short] = s.Original
	}
	return matchPairs
}

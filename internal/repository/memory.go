package repository

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"strconv"
	"sync"

	"github.com/eshadow1/shortener/internal/loggers"
	"github.com/eshadow1/shortener/internal/model"
)

var ErrShortNotFound = errors.New("short not found")

type memoryRepository struct {
	matchPairs  map[string]map[string]string
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

func (m *memoryRepository) Save(ctx context.Context, values []model.URLInfo) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	userID := ctx.Value(model.UserIDContextKey).(string)

	if m.matchPairs[userID] == nil {
		m.matchPairs[userID] = make(map[string]string)
	}

	for _, value := range values {
		m.matchPairs[userID][value.ShortURL] = value.OriginalURL
	}
	return nil
}

func (m *memoryRepository) Get(ctx context.Context, short string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	userID := ctx.Value(model.UserIDContextKey).(string)

	if userInfo, okUser := m.matchPairs[userID]; okUser {
		if val, okShort := userInfo[short]; okShort {
			return val, nil
		}
	}

	return "", ErrShortNotFound
}

func (m *memoryRepository) GetUserURLs(ctx context.Context) ([]model.UserURL, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	userID := ctx.Value(model.UserIDContextKey).(string)
	userInfo, okUser := m.matchPairs[userID]

	urls := make([]model.UserURL, 0, len(userInfo))
	if !okUser {
		return urls, nil
	}

	for short, origin := range userInfo {
		urls = append(urls, model.UserURL{ShortURL: short, OriginalURL: origin})
	}

	return urls, nil
}

func (m *memoryRepository) Close() {
	saveData(m.storagePath, m.matchPairs)
}

func saveData(storagePath string, data map[string]map[string]string) {
	if len(data) == 0 {
		return
	}
	temp := make([]model.FileStorage, 0, len(data))
	index := int64(1)
	for user, info := range data {
		for short, origin := range info {
			temp = append(temp, model.FileStorage{
				UUID:     strconv.FormatInt(index, 10),
				Short:    short,
				Original: origin,
				UserID:   user,
			})
			index++
		}
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

func loadData(storagePath string) map[string]map[string]string {
	matchPairs := make(map[string]map[string]string)
	if storagePath == "" {
		return matchPairs
	}

	data, errRead := os.ReadFile(storagePath)
	if errRead != nil {
		loggers.Log.Errorf("Error reading file %s: %s", storagePath, errRead)
		return matchPairs
	}

	if len(data) == 0 {
		return matchPairs
	}

	var temp []model.FileStorage
	errUnmarshal := json.Unmarshal(data, &temp)
	if errUnmarshal != nil {
		loggers.Log.Errorf("Error unmarshalling file %s: %s", storagePath, errUnmarshal)
		return matchPairs
	}

	for _, s := range temp {
		matchPairs[s.UserID][s.Short] = s.Original
	}
	return matchPairs
}

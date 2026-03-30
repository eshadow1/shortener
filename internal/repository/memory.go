package repository

import (
	"context"
	"errors"
	"sync"
)

var ErrShortNotFound = errors.New("short not found")

type memoryRepository struct {
	matchPairs map[string]string
	mu         sync.RWMutex
}

func NewMemoryRepository() *memoryRepository {
	return &memoryRepository{
		matchPairs: make(map[string]string),
		mu:         sync.RWMutex{},
	}
}

func (m *memoryRepository) Save(_ context.Context, short, origin string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.matchPairs[short] = origin
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

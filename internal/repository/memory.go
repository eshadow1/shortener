package repository

import (
	"fmt"
	"sync"
)

type memoryRepository struct {
	m  map[string]string
	mx sync.Mutex
}

func NewMemoryRepository() *memoryRepository {
	return &memoryRepository{
		m:  make(map[string]string),
		mx: sync.Mutex{},
	}
}

func (m *memoryRepository) Save(short, origin string) error {
	m.mx.Lock()
	defer m.mx.Unlock()

	m.m[short] = origin
	return nil
}

func (m *memoryRepository) Get(short string) (string, error) {
	m.mx.Lock()
	defer m.mx.Unlock()

	if val, ok := m.m[short]; ok {
		return val, nil
	}

	return "", fmt.Errorf("short not found")
}

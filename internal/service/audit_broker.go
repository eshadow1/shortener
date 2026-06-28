package service

import (
	"sync"

	"github.com/eshadow1/shortener/internal/model"
)

// Observer — интерфейс наблюдателя
type Observer interface {
	Notify(event model.Event)
}

// AuditBroker — субъект, который хранит наблюдателей и рассылает события
type AuditBroker struct {
	mu        sync.RWMutex
	observers []Observer
}

func NewAuditBroker() *AuditBroker {
	return &AuditBroker{
		observers: make([]Observer, 0),
	}
}

// Register добавляет наблюдателя в список
func (b *AuditBroker) Register(observer Observer) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.observers = append(b.observers, observer)
}

// Notify рассылает событие всем зарегистрированным наблюдателям
func (b *AuditBroker) Notify(event model.Event) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	for _, observer := range b.observers {
		go observer.Notify(event)
	}
}

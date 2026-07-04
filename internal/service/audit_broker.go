package service

import (
	"sync"

	"github.com/eshadow1/shortener/internal/model"
)

const (
	// DefaultWorkerCount — количество воркеров по умолчанию
	DefaultWorkerCount = 5

	// DefaultTaskBufferSize — размер буфера канала задач
	DefaultTaskBufferSize = 1000
)

// Observer — интерфейс наблюдателя
type Observer interface {
	Notify(event model.Event)
	Close()
}

// task — задача для воркера
type task struct {
	observer Observer
	event    model.Event
}

// AuditBroker — субъект, который хранит наблюдателей и рассылает события
type AuditBroker struct {
	mu        sync.RWMutex
	observers []Observer
	taskCh    chan task
	wg        sync.WaitGroup
}

// NewAuditBroker создает и возвращает новый экземпляр для работы с рассылкой событий наблюдателям.
func NewAuditBroker() *AuditBroker {
	b := &AuditBroker{
		observers: make([]Observer, 0),
		taskCh:    make(chan task, DefaultTaskBufferSize),
	}

	b.startWorkers(DefaultWorkerCount)

	return b
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
		b.taskCh <- task{observer: observer, event: event}
	}
}

// Close закрывает канал задач и ожидает завершения всех воркеров
func (b *AuditBroker) Close() {
	close(b.taskCh)
	b.wg.Wait()

	// Закрываем наблюдателей
	b.mu.RLock()
	defer b.mu.RUnlock()
	for _, observer := range b.observers {
		observer.Close()
	}
}

func (b *AuditBroker) startWorkers(workerCount int) {
	b.wg.Add(workerCount)
	for range workerCount {
		go b.worker()
	}
}

func (b *AuditBroker) worker() {
	defer b.wg.Done()
	for t := range b.taskCh {
		t.observer.Notify(t.event)
	}
}

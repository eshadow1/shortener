// Package Pool предоставляет реализацию для повторного использования обьектов
package pool

import "sync"

// Pool — контейнер, который оборачивает sync.Pool.
// T ограничен интерфейсом, требующим наличия метода Reset().
type Pool[T interface{ Reset() }] struct {
	p       sync.Pool
	creator func() T
}

// New — конструктор, создающая и возвращающая указатель на новый Pool.
func New[T interface{ Reset() }](creator func() T) *Pool[T] {
	return &Pool[T]{
		creator: creator,
	}
}

// Get возвращает объект из пула.
// Если пул пуст, возвращается нулевое значение типа T.
func (p *Pool[T]) Get() T {
	v := p.p.Get()
	if v == nil {
		if p.creator != nil {
			return p.creator()
		}
		var zero T
		return zero
	}
	return v.(T)
}

// Put выполняет сброс состояния объекта и помещает его обратно в пул.
func (p *Pool[T]) Put(obj T) {
	obj.Reset()
	p.p.Put(obj)
}

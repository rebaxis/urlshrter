// Package lib содержит вспомогательные функции и структуры для сервиса сокращения URL.
package lib

import "sync"

// Resetable - интерфейс для типов с методом Reset().
type Resetable interface {
	Reset()
}

// Pool - генерик-пул объектов с автоматическим вызовом Reset() при возврате в пул.
type Pool[T Resetable] struct {
	pool *sync.Pool
}

// New создаёт новый Pool для типа T.
// newFunc - фабричная функция для создания новых экземпляров.
func New[T Resetable](newFunc func() T) *Pool[T] {
	return &Pool[T]{
		pool: &sync.Pool{
			New: func() interface{} {
				return newFunc()
			},
		},
	}
}

// Get возвращает объект из пула или создаёт новый, если пул пуст.
func (p *Pool[T]) Get() T {
	return p.pool.Get().(T)
}

// Put возвращает объект в пул после автоматического вызова Reset().
func (p *Pool[T]) Put(obj T) {
	obj.Reset()
	p.pool.Put(obj)
}

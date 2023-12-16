package queue

import (
	"errors"
	"sync"
)

type ErrEmptyQueue error

type Queue[T any] struct {
	mu  sync.RWMutex
	buf []T
}

func New[T any](capacity int) Queue[T] {
	return Queue[T]{
		buf: make([]T, 0, capacity),
	}
}

func (q *Queue[T]) IsEmpty() bool {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return len(q.buf) == 0
}

func (q *Queue[T]) Size() int {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return len(q.buf)
}

func (q *Queue[T]) Enqueue(value T) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.buf = append(q.buf, value)
}

func (q *Queue[T]) Dequeue() (T, error) {
	if q.IsEmpty() {
		return *new(T), ErrEmptyQueue(errors.New("queue is empty"))
	}

	q.mu.Lock()
	defer q.mu.Unlock()

	value := q.buf[0]
	q.buf = q.buf[1:]
	return value, nil
}

package queue

import (
	"errors"
	"sync"
)

var ErrEmptyQueue = errors.New("queue is empty")

type Queue[T any] struct {
	mu  sync.RWMutex
	buf []T
}

func NewQueue[T any](capacity int) *Queue[T] {
	return &Queue[T]{
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
	value, err := q.Peek()
	if err != nil {
		return *new(T), err
	}

	if err := q.Pop(); err != nil {
		return *new(T), err
	}

	return value, nil
}

func (q *Queue[T]) Peek() (T, error) {
	if q.IsEmpty() {
		return *new(T), ErrEmptyQueue
	}

	q.mu.RLock()
	defer q.mu.RUnlock()

	value := q.buf[0]
	return value, nil
}

func (q *Queue[T]) Pop() error {
	if q.IsEmpty() {
		return ErrEmptyQueue
	}

	q.mu.Lock()
	defer q.mu.Unlock()

	q.buf = q.buf[1:]
	return nil
}

func (q *Queue[T]) GetAll() []T {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return q.buf
}

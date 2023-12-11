package queue

import (
	"errors"
)

type ErrEmptyQueue error

// TODO: Should be unit-tested
type Queue[T any] struct {
	buf []T
}

func New[T any](capacity int) Queue[T] {
	return Queue[T]{
		buf: make([]T, capacity),
	}
}

func (q Queue[T]) IsEmpty() bool {
	return len(q.buf) == 0
}

func (q Queue[T]) Size() int {
	return len(q.buf)
}

func (q *Queue[T]) Enqueue(value T) {
	q.buf = append(q.buf, value)
}

func (q *Queue[T]) Dequeue() (T, error) {
	if q.IsEmpty() {
		return *new(T), ErrEmptyQueue(errors.New("queue is empty"))
	}

	value := q.buf[0]
	q.buf = q.buf[1:]
	return value, nil
}

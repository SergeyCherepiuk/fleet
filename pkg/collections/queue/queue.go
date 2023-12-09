package queue

import (
	"errors"
)

type EmptyQueue error

// TODO: Should be unit-tested
type Queue[T any] struct {
	buf []T
}

func (q Queue[T]) IsEmpty() bool {
	return len(q.buf) == 0
}

func (q *Queue[T]) Enqueue(value T) {
	q.buf = append(q.buf, value)
}

func (q *Queue[T]) Dequeue() (T, error) {
	if q.IsEmpty() {
		return *new(T), EmptyQueue(errors.New("queue is empty"))
	}

	value := q.buf[0]
	q.buf = q.buf[1:]
	return value, nil
}

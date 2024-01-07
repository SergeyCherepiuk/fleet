package queue

import (
	"sync"
	"time"

	"github.com/google/uuid"
	"golang.org/x/exp/maps"
)

type TimeBasedQueue[T any] struct {
	mu  sync.RWMutex
	buf map[uuid.UUID]T
	out chan T
}

func NewTimeBasedQueue[T any](internal time.Duration) *TimeBasedQueue[T] {
	return &TimeBasedQueue[T]{
		buf: make(map[uuid.UUID]T),
		out: make(chan T),
	}
}

func (tbq *TimeBasedQueue[T]) Out() <-chan T {
	return tbq.out
}

func (tbq *TimeBasedQueue[T]) GetAll() []T {
	tbq.mu.RLock()
	defer tbq.mu.RUnlock()
	return maps.Values(tbq.buf)
}

func (tbq *TimeBasedQueue[T]) EnqueueNow(value T) {
	id := tbq.put(value)
	go func() {
		tbq.out <- value
		tbq.delete(id)
	}()
}

func (tbq *TimeBasedQueue[T]) EnqueueWithDelay(delay time.Duration, value T) {
	id := tbq.put(value)
	go func() {
		time.Sleep(delay)
		tbq.out <- value
		tbq.delete(id)
	}()
}

func (tbq *TimeBasedQueue[T]) Close() {
	close(tbq.out)
}

func (tbq *TimeBasedQueue[T]) put(value T) uuid.UUID {
	id := uuid.New()
	tbq.mu.Lock()
	tbq.buf[id] = value
	tbq.mu.Unlock()
	return id
}

func (tbq *TimeBasedQueue[T]) delete(id uuid.UUID) {
	tbq.mu.Lock()
	delete(tbq.buf, id)
	tbq.mu.Unlock()
}

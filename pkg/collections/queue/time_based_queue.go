package queue

import (
	"sync"
	"time"

	"golang.org/x/exp/maps"
)

type TimeBasedQueue[T any] struct {
	mu  sync.RWMutex
	buf map[time.Time]T

	interval time.Duration
	out      chan T
}

func NewTimeBasedQueue[T any](bufferSize int, internal time.Duration) *TimeBasedQueue[T] {
	var out chan T
	if bufferSize == 0 {
		out = make(chan T)
	} else {
		out = make(chan T, bufferSize)
	}

	tbq := TimeBasedQueue[T]{
		buf:      make(map[time.Time]T),
		interval: internal,
		out:      out,
	}

	go tbq.Watch()

	return &tbq
}

func (tbq *TimeBasedQueue[T]) Out() <-chan T {
	return tbq.out
}

func (tbq *TimeBasedQueue[T]) GetAll() []T {
	tbq.mu.RLock()
	defer tbq.mu.RUnlock()
	return maps.Values(tbq.buf)
}

func (tbq *TimeBasedQueue[T]) Watch() {
	for range time.NewTicker(tbq.interval).C {
		now := time.Now()
		tbq.mu.Lock()
		for processAfter, value := range tbq.buf {
			if processAfter.Before(now) {
				tbq.out <- value
				delete(tbq.buf, processAfter)
			}
		}
		tbq.mu.Unlock()
	}
}

func (tbq *TimeBasedQueue[T]) Enqueue(value T) {
	tbq.mu.Lock()
	defer tbq.mu.Unlock()
	tbq.buf[time.Now()] = value
}

func (tbq *TimeBasedQueue[T]) EnqueueWithDelay(processAfter time.Time, value T) {
	tbq.mu.Lock()
	defer tbq.mu.Unlock()
	tbq.buf[processAfter] = value
}

func (tbq *TimeBasedQueue[T]) Close() {
	close(tbq.out)
}

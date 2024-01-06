package queue

import (
	"context"
	"sync"
	"time"

	mapsinternal "github.com/SergeyCherepiuk/fleet/internal/maps"
	"golang.org/x/exp/maps"
)

type TimeBasedQueue[T any] struct {
	mu  sync.RWMutex
	buf map[time.Time]T

	interval time.Duration
	out      chan T

	cancel context.CancelFunc
}

func NewTimeBasedQueue[T any](internal time.Duration) *TimeBasedQueue[T] {
	ctx, cancel := context.WithCancel(context.Background())
	tbq := TimeBasedQueue[T]{
		buf:      make(map[time.Time]T),
		interval: internal,
		out:      make(chan T),
		cancel:   cancel,
	}

	go tbq.Watch(ctx)

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

func (tbq *TimeBasedQueue[T]) Watch(ctx context.Context) {
	for range time.Tick(tbq.interval) {
		select {
		case <-ctx.Done():
			return

		default:
			for processAfter, value := range mapsinternal.ConcurrentCopy(tbq.buf) {
				if processAfter.Before(time.Now()) {
					tbq.out <- value

					tbq.mu.Lock()
					delete(tbq.buf, processAfter)
					tbq.mu.Unlock()
				}
			}
		}
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
	tbq.cancel()
	close(tbq.out)
}

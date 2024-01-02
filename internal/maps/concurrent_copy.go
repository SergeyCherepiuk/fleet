package maps

import "sync"

var muConcurrentCopy sync.RWMutex

func ConcurrentCopy[K comparable, V any](m map[K]V) map[K]V {
	muConcurrentCopy.RLock()
	defer muConcurrentCopy.RUnlock()

	c := make(map[K]V, len(m))
	for k, v := range m {
		c[k] = v
	}
	return c
}

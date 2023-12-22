package scheduler

import (
	"github.com/SergeyCherepiuk/fleet/pkg/registry"
	"github.com/SergeyCherepiuk/fleet/pkg/task"
)

type RoundRobin struct {
	last int
}

func (s *RoundRobin) SelectWorker(t task.Task, wes []registry.WorkerEntry) (registry.WorkerEntry, error) {
	if len(wes) == 0 {
		return registry.WorkerEntry{}, ErrNoWorkersAvailable
	}

	if s.last+1 < len(wes) {
		s.last++
	} else {
		s.last = 0
	}

	return wes[s.last], nil
}

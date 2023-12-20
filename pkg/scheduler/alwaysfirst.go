package scheduler

import (
	"errors"

	"github.com/SergeyCherepiuk/fleet/pkg/registry"
	"github.com/SergeyCherepiuk/fleet/pkg/task"
)

type AlwaysFirst struct{}

func (s *AlwaysFirst) SelectWorker(t task.Task, wes []registry.WorkerEntry) (registry.WorkerEntry, error) {
	if len(wes) > 0 {
		return wes[0], nil
	}

	return registry.WorkerEntry{}, ErrNoWorkersAvailable(errors.New("no workers available"))
}

package scheduler

import (
	"errors"

	"github.com/SergeyCherepiuk/fleet/pkg/registry"
	"github.com/SergeyCherepiuk/fleet/pkg/task"
)

var ErrNoWorkersAvailable = errors.New("no workers available")

type Scheduler interface {
	SelectWorker(task task.Task, workers []registry.WorkerEntry) (registry.WorkerEntry, error)
}

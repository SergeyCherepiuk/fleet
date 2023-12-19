package scheduler

import (
	"github.com/SergeyCherepiuk/fleet/pkg/registry"
	"github.com/SergeyCherepiuk/fleet/pkg/task"
)

type ErrNoWorkersAvailable error

type Scheduler interface {
	SelectWorker(task task.Task, workers []registry.WorkerEntry) (registry.WorkerEntry, error)
}

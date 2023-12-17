package scheduler

import (
	"github.com/SergeyCherepiuk/fleet/pkg/manager/registry"
	"github.com/SergeyCherepiuk/fleet/pkg/task"
	"github.com/google/uuid"
)

type ErrNoWorkersAvailable error

type Scheduler interface {
	SelectWorker(task task.Task, workers map[uuid.UUID]registry.Entry) (uuid.UUID, registry.Entry, error)
}

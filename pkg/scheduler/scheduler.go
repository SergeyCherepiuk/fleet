package scheduler

import (
	"errors"

	"github.com/SergeyCherepiuk/fleet/pkg/registry"
	"github.com/SergeyCherepiuk/fleet/pkg/task"
	"github.com/google/uuid"
)

var ErrNoWorkersAvailable = errors.New("no workers available")

type Scheduler interface {
	SelectWorker(task task.Task, workers map[uuid.UUID]registry.WorkerEntry) (uuid.UUID, registry.WorkerEntry, error)
}

package scheduler

import (
	"errors"

	"github.com/SergeyCherepiuk/fleet/pkg/consensus"
	"github.com/SergeyCherepiuk/fleet/pkg/task"
	"github.com/google/uuid"
)

var (
	ErrNoAvailableWorkers = errors.New("no available workers")
	ErrNoCapableWorkers   = errors.New("no capable workers")
)

type Scheduler interface {
	SelectWorker(task task.Task, workers map[uuid.UUID]consensus.Worker) (uuid.UUID, consensus.Worker, error)
}

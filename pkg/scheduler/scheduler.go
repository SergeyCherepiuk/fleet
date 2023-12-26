package scheduler

import (
	"errors"

	"github.com/SergeyCherepiuk/fleet/pkg/consensus"
	"github.com/SergeyCherepiuk/fleet/pkg/task"
	"github.com/google/uuid"
)

var ErrNoWorkersAvailable = errors.New("no workers available")

type Scheduler interface {
	SelectWorker(task task.Task, workers map[uuid.UUID]consensus.Worker) (uuid.UUID, consensus.Worker, error)
}

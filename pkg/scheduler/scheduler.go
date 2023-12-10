package scheduler

import (
	"github.com/SergeyCherepiuk/fleet/pkg/task"
	"github.com/SergeyCherepiuk/fleet/pkg/worker"
)

type Scheduler interface {
	SelectWorker(task task.Task, workers []worker.Worker) (worker.Worker, error)
}

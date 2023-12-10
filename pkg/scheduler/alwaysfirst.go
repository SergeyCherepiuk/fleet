package scheduler

import (
	"errors"

	"github.com/SergeyCherepiuk/fleet/pkg/task"
	"github.com/SergeyCherepiuk/fleet/pkg/worker"
)

type ErrNoWorkersAvailable error

type AlwaysFirst struct{}

func (s AlwaysFirst) SelectWorker(task task.Task, workers []worker.Worker) (worker.Worker, error) {
	if len(workers) == 0 {
		return worker.Worker{}, ErrNoWorkersAvailable(errors.New("no workers available"))
	}
	return workers[0], nil
}

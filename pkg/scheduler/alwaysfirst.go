package scheduler

import (
	"errors"

	"github.com/SergeyCherepiuk/fleet/pkg/node"
	"github.com/SergeyCherepiuk/fleet/pkg/task"
)

type AlwaysFirst struct{}

func (s AlwaysFirst) SelectWorker(task task.Task, workers []node.Addr) (node.Addr, error) {
	if len(workers) == 0 {
		return node.Addr{}, ErrNoWorkersAvailable(errors.New("no workers available"))
	}
	return workers[0], nil
}

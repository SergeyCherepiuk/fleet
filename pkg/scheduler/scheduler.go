package scheduler

import (
	"github.com/SergeyCherepiuk/fleet/pkg/node"
	"github.com/SergeyCherepiuk/fleet/pkg/task"
)

type ErrNoWorkersAvailable error

type Scheduler interface {
	SelectWorker(task task.Task, workers []node.Addr) (node.Addr, error)
}

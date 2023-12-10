package worker

import (
	"errors"

	"github.com/SergeyCherepiuk/fleet/pkg/collections/queue"
	"github.com/SergeyCherepiuk/fleet/pkg/docker"
	"github.com/SergeyCherepiuk/fleet/pkg/task"
	"github.com/google/uuid"
)

type ErrInvalidEventState error

type Worker struct {
	ID    uuid.UUID
	Tasks map[uuid.UUID]task.Task
	Queue queue.Queue[task.Task]
}

func (w *Worker) Execute(event task.Event) error {
	switch event.State {
	case task.Running:
		return w.run(event.Task)
	case task.Finished:
		return w.finish(event.Task)
	}
	return ErrInvalidEventState(errors.New("worker: invalid event state"))
}

func (w *Worker) run(task task.Task) error {
	return docker.Run(task.Container)
}

func (w *Worker) finish(task task.Task) error {
	return nil
}

// TODO: Change return type
func (w Worker) Stats() error {
	return nil
}

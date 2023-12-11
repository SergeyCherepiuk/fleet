package worker

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/SergeyCherepiuk/fleet/pkg/collections/queue"
	"github.com/SergeyCherepiuk/fleet/pkg/docker"
	"github.com/SergeyCherepiuk/fleet/pkg/node"
	"github.com/SergeyCherepiuk/fleet/pkg/task"
	"github.com/google/uuid"
)

type ErrInvalidEventState error

type Worker struct {
	node.Node
	ID    uuid.UUID
	Tasks map[uuid.UUID]task.Task
	Queue queue.Queue[task.Task] // TODO(SergeyCherepiuk): Consider if it's needed
}

func (w *Worker) Execute(event task.Event) error {
	if event.Task.State == event.State {
		return nil
	}

	if !task.IsStateTransitionAllowed(event.Task.State, event.State) {
		err := fmt.Errorf("transition from %q to %q is not allowed", event.Task.State, event.State)
		return task.ErrStateTransitionNotAllowed(err)
	}

	switch event.State {
	case task.Running:
		return w.run(event.Task)
	case task.Finished:
		return w.finish(event.Task.ID) // TODO(SergeyCherepiuk): Using ID is an ugly workaround
	}

	return ErrInvalidEventState(errors.New("worker: invalid event state"))
}

func (w *Worker) run(t task.Task) error {
	defer func() {
		t.StartedAt = time.Now()
		w.Tasks[t.ID] = t
	}()

	ctx := context.Background()
	id, err := docker.Run(ctx, t.Container)
	if err != nil {
		t.State = task.Failed
		return err
	}

	t.Container.ID = id
	t.State = task.Running
	return nil
}

func (w *Worker) finish(taskId uuid.UUID) error {
	t := w.Tasks[taskId]
	defer func() {
		t.FinishedAt = time.Now()
		w.Tasks[taskId] = t
	}()

	ctx := context.Background()
	if err := docker.Stop(ctx, t.Container); err != nil {
		t.State = task.Failed
		return err
	}

	t.State = task.Finished
	return nil
}

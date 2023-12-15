package worker

import (
	"context"
	"fmt"
	"time"

	"github.com/SergeyCherepiuk/fleet/pkg/c14n"
	"github.com/SergeyCherepiuk/fleet/pkg/node"
	"github.com/SergeyCherepiuk/fleet/pkg/task"
	"github.com/google/uuid"
)

type ErrInvalidEventState error

type Worker struct {
	node.Node
	c14n.Runtime
	ID    uuid.UUID
	Tasks map[uuid.UUID]task.Task
}

func (w *Worker) Run(t task.Task) error {
	defer func() {
		t.StartedAt = time.Now()
		w.Tasks[t.ID] = t
		fmt.Printf("%+v\n", w.Tasks)
	}()

	ctx := context.Background()
	id, err := w.Runtime.Run(ctx, t.Container)
	if err != nil {
		t.State = task.Failed
		return err
	}

	t.Container.ID = id
	t.State = task.Running
	return nil
}

func (w *Worker) Finish(taskId uuid.UUID) error {
	t := w.Tasks[taskId]
	defer func() {
		t.FinishedAt = time.Now()
		w.Tasks[taskId] = t
	}()

	ctx := context.Background()
	if err := w.Runtime.Stop(ctx, t.Container); err != nil {
		t.State = task.Failed
		return err
	}

	t.State = task.Finished
	return nil
}

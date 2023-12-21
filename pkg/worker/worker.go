package worker

import (
	"context"
	"time"

	"github.com/SergeyCherepiuk/fleet/pkg/c14n"
	"github.com/SergeyCherepiuk/fleet/pkg/node"
	"github.com/SergeyCherepiuk/fleet/pkg/task"
	"github.com/google/uuid"
)

const HeartbeatInterval = time.Second * 10

type ErrInvalidEventState error

type Worker struct {
	ID      uuid.UUID
	node    node.Node
	runtime c14n.Runtime
}

func New(node node.Node, runtime c14n.Runtime) *Worker {
	return &Worker{
		ID:      uuid.New(),
		node:    node,
		runtime: runtime,
	}
}

func (w *Worker) Run(ctx context.Context, t *task.Task) error {
	defer func() { t.StartedAt = time.Now() }()

	id, err := w.runtime.Run(ctx, t.Container)
	if err != nil {
		// TODO(SergeyCherepiuk): Put the error on the event/message bus
		t.State = task.Failed
		return err
	}

	t.Container.ID = id
	t.State = task.Running
	return nil
}

func (w *Worker) Finish(ctx context.Context, t *task.Task) error {
	defer func() { t.FinishedAt = time.Now() }()

	if err := w.runtime.Stop(ctx, t.Container); err != nil {
		t.State = task.Failed
		return err
	}

	t.State = task.Finished
	return nil
}

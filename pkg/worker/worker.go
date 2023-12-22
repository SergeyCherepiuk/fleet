package worker

import (
	"context"
	"fmt"
	"time"

	"github.com/SergeyCherepiuk/fleet/pkg/c14n"
	"github.com/SergeyCherepiuk/fleet/pkg/httpclient"
	"github.com/SergeyCherepiuk/fleet/pkg/node"
	"github.com/SergeyCherepiuk/fleet/pkg/task"
	"github.com/google/uuid"
)

const HeartbeatInterval = time.Second * 10

type Worker struct {
	ID          uuid.UUID
	node        node.Node
	runtime     c14n.Runtime
	managerAddr string
}

func New(node node.Node, runtime c14n.Runtime, managerAddr string) *Worker {
	worker := &Worker{
		ID:          uuid.New(),
		node:        node,
		runtime:     runtime,
		managerAddr: managerAddr,
	}

	go worker.registerWorker()
	go worker.sendHeartbeats()

	return worker
}

func (w *Worker) Run(ctx context.Context, t *task.Task) error {
	defer func() {
		t.StartedAt = time.Now()
		message := Message{From: w.ID, Task: *t}
		httpclient.Post(w.managerAddr, "/worker/message", message)
	}()

	id, err := w.runtime.Run(ctx, t.Container)
	if err != nil {
		t.State = task.Failed
		return err
	}

	t.Container.ID = id
	t.State = task.Running
	return nil
}

func (w *Worker) Finish(ctx context.Context, t *task.Task) error {
	defer func() {
		t.FinishedAt = time.Now()
		message := Message{From: w.ID, Task: *t}
		httpclient.Post(w.managerAddr, "/worker/message", message)
	}()

	if err := w.runtime.Stop(ctx, t.Container); err != nil {
		t.State = task.Failed
		return err
	}

	t.State = task.Finished
	return nil
}

func (w *Worker) registerWorker() error {
	endpoint := fmt.Sprintf("/worker/%s", w.ID)
	_, err := httpclient.Post(w.managerAddr, endpoint, w.node.Addr)
	return err
}

func (w *Worker) sendHeartbeats() {
	for {
		endpoint := fmt.Sprintf("/worker/%s", w.ID)
		httpclient.Put(w.managerAddr, endpoint, nil)
		time.Sleep(HeartbeatInterval)
	}
}

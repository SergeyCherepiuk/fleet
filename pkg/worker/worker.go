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

const (
	HeartbeatInterval = time.Second * 10
	InspectInterval   = time.Second
)

// TODO(SergeyCherepiuk): Worker should periodically query the list of
// running containers to make sure none of the task failed
type Worker struct {
	Id          uuid.UUID
	node        node.Node
	runtime     c14n.Runtime
	managerAddr string
}

func New(node node.Node, runtime c14n.Runtime, managerAddr string) *Worker {
	worker := &Worker{
		Id:          uuid.New(),
		node:        node,
		runtime:     runtime,
		managerAddr: managerAddr,
	}

	worker.registerWorker()
	go worker.sendHeartbeats()
	go worker.inspectTasks()

	return worker
}

func (w *Worker) Run(ctx context.Context, t *task.Task) error {
	defer func() {
		t.StartedAt = time.Now()
		message := Message{From: w.Id, Task: *t}
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
		message := Message{From: w.Id, Task: *t}
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
	endpoint := fmt.Sprintf("/worker/%s", w.Id)
	_, err := httpclient.Post(w.managerAddr, endpoint, w.node.Addr)
	return err
}

func (w *Worker) sendHeartbeats() {
	for {
		endpoint := fmt.Sprintf("/worker/%s", w.Id)
		httpclient.Put(w.managerAddr, endpoint, nil)
		time.Sleep(HeartbeatInterval)
	}
}

func (w *Worker) inspectTasks() {
	for {
		ctx := context.Background()
		containers, err := w.runtime.Containers(ctx)
		if err != nil {
			continue
		}

		containerIdsToStates := make(map[string]string)
		for _, container := range containers {
			state, _ := w.runtime.ContainerState(ctx, container.ID)
			containerIdsToStates[container.ID] = state
		}

		// TODO(SergeyCherepiuk): After worker will also own its copy of the registry
		// compute the difference between desired and actual states and report is to the manager

		time.Sleep(InspectInterval)
	}
}

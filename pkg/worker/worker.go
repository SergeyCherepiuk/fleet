package worker

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"time"

	"github.com/SergeyCherepiuk/fleet/pkg/c14n"
	"github.com/SergeyCherepiuk/fleet/pkg/collections/queue"
	"github.com/SergeyCherepiuk/fleet/pkg/consensus"
	"github.com/SergeyCherepiuk/fleet/pkg/container"
	"github.com/SergeyCherepiuk/fleet/pkg/httpclient"
	"github.com/SergeyCherepiuk/fleet/pkg/node"
	"github.com/SergeyCherepiuk/fleet/pkg/task"
	"github.com/google/uuid"
)

const (
	InspectInterval        = time.Second
	ShutdownTimeoutSeconds = 5
)

// TODO(SergeyCherepiuk): Worker should periodically query the list of
// running containers to make sure none of the task failed
type Worker struct {
	Id           uuid.UUID
	node         node.Node
	runtime      c14n.Runtime
	store        consensus.Store
	managerAddr  string
	shutdownCmds queue.Queue[*exec.Cmd]
}

func New(node node.Node, runtime c14n.Runtime, managerAddr string) *Worker {
	worker := &Worker{
		Id:           uuid.New(),
		node:         node,
		runtime:      runtime,
		store:        consensus.NewLocalStore(),
		managerAddr:  managerAddr,
		shutdownCmds: queue.New[*exec.Cmd](0),
	}

	worker.register()
	go worker.inspectTasks()
	go worker.spawnShutdownProcesses()
	go worker.catchInterrupt()

	go worker.inspectStore()

	return worker
}

func (w *Worker) inspectStore() {
	for {
		for id, worker := range w.store.AllWorkers() {
			fmt.Println(id, worker.Tasks)
		}
		time.Sleep(time.Second)
	}
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

func (w *Worker) CommitChanges(cmds ...consensus.Command) (int, error) {
	for _, cmd := range cmds {
		if off, err := w.store.CommitChange(cmd); err != nil {
			return off, err
		}
	}
	return 0, nil
}

func (w *Worker) CancleShutdown() error {
	cmd, err := w.shutdownCmds.Dequeue()
	if err != nil {
		return nil
	}

	if cmd == nil || cmd.Process == nil {
		return nil
	}

	return exec.Command("kill", "-9", fmt.Sprint(cmd.Process.Pid)).Run()
}

func (w *Worker) register() error {
	endpoint := fmt.Sprintf("/worker/%s", w.Id)
	_, err := httpclient.Post(w.managerAddr, endpoint, w.node.Addr)
	return err
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

func (w *Worker) spawnShutdownProcesses() {
	for {
		sleep := fmt.Sprintf("sleep %d", ShutdownTimeoutSeconds)
		kill := fmt.Sprintf("kill -15 %d", os.Getpid())
		stop := fmt.Sprintf(
			"docker rm -f $(docker ps -qaf 'label=%s=%s')",
			container.TypeLabelKey, container.TypeLabelValue,
		)

		commands := strings.Join([]string{sleep, kill, stop}, "; ")
		cmd := exec.Command("/bin/sh", "-c", commands)
		w.shutdownCmds.Enqueue(cmd)
		cmd.Run()
	}
}

func (w *Worker) catchInterrupt() {
	ch := make(chan os.Signal)
	signal.Notify(ch, os.Interrupt)
	<-ch

	defer os.Exit(0)

	ctx := context.Background()
	containers, err := w.runtime.Containers(ctx)
	if err != nil {
		return
	}

	for _, container := range containers {
		w.runtime.Stop(ctx, container)
	}
}

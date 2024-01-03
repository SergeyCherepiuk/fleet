package worker

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"time"

	mapsinternal "github.com/SergeyCherepiuk/fleet/internal/maps"
	"github.com/SergeyCherepiuk/fleet/pkg/c14n"
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

type Worker struct {
	Id           uuid.UUID
	Node         node.Node
	runtime      c14n.Runtime
	store        consensus.Store
	managerAddr  string
	shutdownCmds chan *exec.Cmd
}

type Message struct {
	From uuid.UUID
	Task task.Task
}

func New(node node.Node, runtime c14n.Runtime, managerAddr string) *Worker {
	worker := &Worker{
		Id:           uuid.New(),
		Node:         node,
		runtime:      runtime,
		store:        consensus.NewLocalStore(),
		managerAddr:  managerAddr,
		shutdownCmds: make(chan *exec.Cmd),
	}

	worker.register()
	go worker.inspectTasks()
	go worker.spawnShutdownProcesses()
	go worker.catchInterrupt()

	return worker
}

func (w *Worker) Run(ctx context.Context, t task.Task) error {
	defer func() {
		t.StartedAt = time.Now()
		message := Message{From: w.Id, Task: t}
		httpclient.Post(w.managerAddr, "/worker/message", message)
	}()

	id, err := w.runtime.CreateAndRun(ctx, t.Container)
	if err != nil {
		t.State = task.FailedOnStartup
		return err
	}

	t.Container.Id = id
	t.State = task.Running
	return nil
}

func (w *Worker) Finish(ctx context.Context, t task.Task) error {
	defer func() {
		t.FinishedAt = time.Now()
		message := Message{From: w.Id, Task: t}
		httpclient.Post(w.managerAddr, "/worker/message", message)
	}()

	if err := w.runtime.StopAndRemove(ctx, t.Container.Id); err != nil {
		t.State = task.FailedAfterStartup
		return err
	}

	t.State = task.Finished
	return nil
}

type Info struct {
	Id          uuid.UUID
	Addr        node.Addr
	ManagerAddr string
	TasksCount  int
	RuntimeName string
}

func (w *Worker) Info() *Info {
	workerFromStore, _ := w.store.GetWorker(w.Id)
	tasksCount := len(workerFromStore.Tasks)

	return &Info{
		Id:          w.Id,
		Addr:        w.Node.Addr,
		ManagerAddr: w.managerAddr,
		TasksCount:  tasksCount,
		RuntimeName: w.runtime.Name(),
	}
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
	cmd := <-w.shutdownCmds
	if cmd == nil || cmd.Process == nil {
		return nil
	}

	return exec.Command("kill", "-9", fmt.Sprint(cmd.Process.Pid)).Run()
}

func (w *Worker) CheckStoreSynchronization(lastIndex int) int {
	return max(0, lastIndex-w.store.LastIndex())
}

func (w *Worker) AvailableResources() (node.Resources, error) {
	workerResources, err := w.Node.Resources()
	if err != nil {
		return node.Resources{}, err
	}

	reservedResources, err := w.ReservedResources()
	if err != nil {
		return node.Resources{}, err
	}

	availableMemory := min(
		workerResources.Memory.Available,
		workerResources.Memory.Total-reservedResources.Memory,
	)
	workerResources.Memory.Available = max(availableMemory, 0)

	availableDisk := min(
		workerResources.Disk.Available,
		workerResources.Disk.Total-reservedResources.Disk,
	)
	workerResources.Disk.Available = max(availableDisk, 0)

	return workerResources, nil
}

func (w *Worker) ReservedResources() (container.RequiredResources, error) {
	worker, err := w.store.GetWorker(w.Id)
	if err != nil {
		return container.RequiredResources{}, err
	}

	var resources container.RequiredResources
	for _, t := range worker.Tasks {
		if t.State == task.Running {
			resources.CPU += t.Container.Config.RequiredResources.CPU
			resources.Memory += t.Container.Config.RequiredResources.Memory
			resources.Disk += t.Container.Config.RequiredResources.Disk
		}
	}
	return resources, nil
}

func (w *Worker) register() error {
	endpoint := fmt.Sprintf("/worker/%s", w.Id)
	_, err := httpclient.Post(w.managerAddr, endpoint, w.Node.Addr)
	return err
}

func mapState(state container.State) task.State {
	switch state {
	case container.State{Status: "created", ExitCode: 0},
		container.State{Status: "running", ExitCode: 0},
		container.State{Status: "restarting", ExitCode: 0}:
		return task.Running

	case container.State{Status: "paused", ExitCode: 0},
		container.State{Status: "exited", ExitCode: 0},
		container.State{Status: "removing", ExitCode: 0}:
		return task.Finished

	default:
		return task.FailedAfterStartup
	}
}

func (w *Worker) inspectTasks() {
	ctx := context.Background()
	for {
		containers, err := w.runtime.Containers(ctx)
		if err != nil {
			continue
		}

		containerIdsToStates := make(map[string]task.State)
		for _, container := range containers {
			state, _ := w.runtime.ContainerState(ctx, container.Id)
			containerIdsToStates[container.Id] = mapState(state)
		}

		worker, err := w.store.GetWorker(w.Id)
		if err != nil {
			continue
		}

		for _, t := range mapsinternal.ConcurrentCopy(worker.Tasks) {
			actualState, ok := containerIdsToStates[t.Container.Id]
			if !ok && t.State == task.Running {
				t.State = task.FailedAfterStartup
				message := Message{From: w.Id, Task: t}
				httpclient.Post(w.managerAddr, "/worker/message", message)
			}

			if !ok {
				continue
			}

			if t.State != actualState {
				t.State = actualState
				message := Message{From: w.Id, Task: t}
				httpclient.Post(w.managerAddr, "/worker/message", message)
			}

			if actualState == task.Finished || actualState.Fail() {
				w.runtime.StopAndRemove(ctx, t.Container.Id)
			}
		}
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
		cmd.Start()
		w.shutdownCmds <- cmd
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
		w.runtime.StopAndRemove(ctx, container.Id)
	}
}

package manager

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	internalhttp "github.com/SergeyCherepiuk/fleet/internal/http"
	"github.com/SergeyCherepiuk/fleet/pkg/manager/registry"
	"github.com/SergeyCherepiuk/fleet/pkg/node"
	"github.com/SergeyCherepiuk/fleet/pkg/scheduler"
	"github.com/SergeyCherepiuk/fleet/pkg/task"
	"github.com/SergeyCherepiuk/fleet/pkg/worker"
	"github.com/google/uuid"
)

type Manager struct {
	id             uuid.UUID
	node           *node.Node
	scheduler      scheduler.Scheduler
	workerRegistry *registry.WorkerRegistry
}

func New(node node.Node, scheduler scheduler.Scheduler) *Manager {
	manager := Manager{
		id:             uuid.New(),
		node:           &node,
		scheduler:      scheduler,
		workerRegistry: registry.New(),
	}
	go manager.workerRegistry.Watch()
	return &manager
}

func (m *Manager) Run(t task.Task) error {
	var err error

	workers := m.workerRegistry.GetAll()
	id, w, err := m.scheduler.SelectWorker(t, workers)
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			t.State = task.Failed
		}
		err = m.workerRegistry.AddTask(id, t)
	}()

	t.State = task.Scheduled
	if err = m.workerRegistry.AddTask(id, t); err != nil {
		return err
	}

	marshaledTask, err := json.Marshal(t)
	if err != nil {
		return err
	}

	workerAddrStr := fmt.Sprintf("%s:%d", w.Addr.Addr, w.Addr.Port)
	url, err := url.JoinPath("http://", workerAddrStr, worker.TaskRunEndpoint)
	body := bytes.NewReader(marshaledTask)

	resp, err := http.Post(url, "application/json", body)
	if err != nil {
		return err
	}

	err = internalhttp.Body(resp, &t)
	return err
}

func (m *Manager) Finish(taskId uuid.UUID) error {
	id, w, err := m.workerRegistry.FindWorker(taskId)
	if err != nil {
		return err
	}

	workerAddr := fmt.Sprintf("%s:%d", w.Addr.Addr, w.Addr.Port)
	url, err := url.JoinPath("http://", workerAddr, worker.TaskFinishEndpoint, taskId.String())

	resp, err := http.Post(url, "", nil)
	if err != nil {
		return err
	}

	var t task.Task
	if err := internalhttp.Body(resp, &t); err != nil {
		return err
	}
	m.workerRegistry.AddTask(id, t)

	return nil
}

package manager

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	internalhttp "github.com/SergeyCherepiuk/fleet/internal/http"
	"github.com/SergeyCherepiuk/fleet/pkg/node"
	"github.com/SergeyCherepiuk/fleet/pkg/registry"
	"github.com/SergeyCherepiuk/fleet/pkg/scheduler"
	"github.com/SergeyCherepiuk/fleet/pkg/task"
	"github.com/SergeyCherepiuk/fleet/pkg/worker"
	"github.com/google/uuid"
	"golang.org/x/exp/maps"
)

type Manager struct {
	id             uuid.UUID
	node           node.Node
	scheduler      scheduler.Scheduler
	workerRegistry registry.WorkerRegistry
}

func New(node node.Node, scheduler scheduler.Scheduler) *Manager {
	manager := Manager{
		id:             uuid.New(),
		node:           node,
		scheduler:      scheduler,
		workerRegistry: make(registry.WorkerRegistry),
	}
	go manager.workerRegistry.Watch()
	return &manager
}

func (m *Manager) Run(t task.Task) error {
	var err error

	workerEntries := m.workerRegistry.GetAll()
	workerEntry, err := m.scheduler.SelectWorker(t, workerEntries)
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			t.State = task.Failed
		}
		workerEntry.Tasks.Add(t)
		err = m.workerRegistry.Set(workerEntry.ID, workerEntry)
	}()

	t.State = task.Scheduled
	workerEntry.Tasks.Add(t)
	if err = m.workerRegistry.Set(workerEntry.ID, workerEntry); err != nil {
		return err
	}

	marshaledTask, err := json.Marshal(t)
	if err != nil {
		return err
	}

	addr := fmt.Sprintf("%s:%d", workerEntry.Addr.Addr, workerEntry.Addr.Port)
	url, err := url.JoinPath("http://", addr, worker.TaskRunEndpoint)
	body := bytes.NewReader(marshaledTask)

	resp, err := http.Post(url, "application/json", body)
	if err != nil {
		return err
	}

	err = internalhttp.Body(resp, &t)
	return err
}

func (m *Manager) Finish(tid uuid.UUID) error {
	workerEntry, err := m.workerRegistry.GetByTaskId(tid)
	if err != nil {
		return err
	}

	t, err := workerEntry.Tasks.Get(tid)
	if err != nil {
		return err
	}

	body, err := json.Marshal(t)
	if err != nil {
		return err
	}

	addr := fmt.Sprintf("%s:%d", workerEntry.Addr.Addr, workerEntry.Addr.Port)
	url, err := url.JoinPath("http://", addr, worker.TaskFinishEndpoint)

	resp, err := http.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}

	if err := internalhttp.Body(resp, &t); err != nil {
		return err
	}

	workerEntry.Tasks.Add(t)
	m.workerRegistry.Set(workerEntry.ID, workerEntry)
	return nil
}

func (m *Manager) Tasks() map[uuid.UUID][]task.Task {
	stat := make(map[uuid.UUID][]task.Task)
	for id, workerEntry := range m.workerRegistry {
		stat[id] = maps.Values(workerEntry.Tasks)
	}
	return stat
}

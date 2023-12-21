package manager

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/SergeyCherepiuk/fleet/pkg/collections/queue"
	internalhttp "github.com/SergeyCherepiuk/fleet/pkg/internal/http"
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
	eventsQueue    queue.Queue[task.Event]
}

func New(node node.Node, scheduler scheduler.Scheduler) *Manager {
	manager := Manager{
		id:             uuid.New(),
		node:           node,
		scheduler:      scheduler,
		workerRegistry: registry.NewWorkerRegistry(),
		eventsQueue:    queue.New[task.Event](0),
	}

	go manager.watchingEventsQueue(100 * time.Millisecond)

	return &manager
}

func (m *Manager) Run(t task.Task) {
	e := task.Event{Task: t, Desired: task.Running}
	m.eventsQueue.Enqueue(e)
}

func (m *Manager) Stop(tid uuid.UUID) error {
	we, err := m.workerRegistry.GetByTaskId(tid)
	if err != nil {
		return err
	}

	t, err := we.Tasks.Get(tid)
	if err != nil {
		return err
	}

	e := task.Event{Task: t, Desired: task.Finished}
	m.eventsQueue.Enqueue(e)
	return nil
}

func (m *Manager) Tasks() map[uuid.UUID][]task.Task {
	stat := make(map[uuid.UUID][]task.Task)
	for id, workerEntry := range m.workerRegistry {
		stat[id] = maps.Values(workerEntry.Tasks)
	}
	return stat
}

func (m *Manager) watchingEventsQueue(interval time.Duration) {
	for {
		event, err := m.eventsQueue.Dequeue()
		if err != nil {
			time.Sleep(interval)
			continue
		}

		switch event.Desired {
		case task.Running:
			m.run(event.Task)
		case task.Finished:
			m.finish(event.Task)
		}
	}
}

func (m *Manager) run(t task.Task) error {
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

func (m *Manager) finish(t task.Task) error {
	workerEntry, err := m.workerRegistry.GetByTaskId(t.ID)
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

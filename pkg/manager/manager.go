package manager

import (
	"fmt"
	"time"

	"github.com/SergeyCherepiuk/fleet/pkg/collections/queue"
	"github.com/SergeyCherepiuk/fleet/pkg/httpclient"
	"github.com/SergeyCherepiuk/fleet/pkg/node"
	"github.com/SergeyCherepiuk/fleet/pkg/registry"
	"github.com/SergeyCherepiuk/fleet/pkg/scheduler"
	"github.com/SergeyCherepiuk/fleet/pkg/task"
	"github.com/SergeyCherepiuk/fleet/pkg/worker"
	"github.com/google/uuid"
	"golang.org/x/exp/maps"
)

const RestartSleepTimeCoefficient = 2

type Manager struct {
	id             uuid.UUID
	node           node.Node
	scheduler      scheduler.Scheduler
	workerRegistry registry.WorkerRegistry
	eventsQueue    queue.Queue[task.Event]
	messagesQueue  queue.Queue[worker.Message]
}

func New(node node.Node, scheduler scheduler.Scheduler) *Manager {
	manager := Manager{
		id:             uuid.New(),
		node:           node,
		scheduler:      scheduler,
		workerRegistry: registry.NewWorkerRegistry(),
		eventsQueue:    queue.New[task.Event](0),
		messagesQueue:  queue.New[worker.Message](0),
	}

	go manager.watchEventsQueue(100 * time.Millisecond)
	go manager.watchMessagesQueue(100 * time.Millisecond)

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

func (m *Manager) watchEventsQueue(interval time.Duration) {
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
		case task.Restarting:
			// TODO(SergeyCherepiuk): Schedule restart if the failure cause is
			// related to image pulling for example, otherwise restart immediately
			// TODO(SergeyCherepiuk): Disregard number of restarts if task is
			// running successfully long enough
			m.scheduleRestart(event.Task)
		}
	}
}

func (m *Manager) watchMessagesQueue(interval time.Duration) {
	for {
		message, err := m.messagesQueue.Dequeue()
		if err != nil {
			time.Sleep(interval)
			continue
		}

		switch message.Task.State {
		case task.Running, task.Finished:
			m.workerRegistry.SetTask(message.From, message.Task)
		case task.Failed:
			message.Task.Restarts = append(message.Task.Restarts, time.Now())
			m.workerRegistry.SetTask(message.From, message.Task)
			event := task.Event{Task: message.Task, Desired: task.Restarting}
			m.eventsQueue.Enqueue(event)
		}
	}
}

func (m *Manager) run(t task.Task) error {
	workerEntries := m.workerRegistry.GetAll()
	workerEntry, err := m.scheduler.SelectWorker(t, workerEntries)
	if err != nil {
		return err
	}

	t.State = task.Scheduled

	addr := fmt.Sprintf("%s:%d", workerEntry.Addr.Addr, workerEntry.Addr.Port)
	httpclient.Post(addr, "/task/run", t)
	return nil
}

func (m *Manager) finish(t task.Task) error {
	workerEntry, err := m.workerRegistry.GetByTaskId(t.ID)
	if err != nil {
		return err
	}

	addr := fmt.Sprintf("%s:%d", workerEntry.Addr.Addr, workerEntry.Addr.Port)
	httpclient.Post(addr, "/task/stop", t)
	return nil
}

func (m *Manager) scheduleRestart(t task.Task) {
	var sleepTime time.Duration
	if len(t.Restarts) < 2 {
		sleepTime = time.Second
	} else {
		l := len(t.Restarts)
		lastSleepTime := t.Restarts[l-1].Sub(t.Restarts[l-2])
		sleepTime = lastSleepTime * RestartSleepTimeCoefficient
	}

	go func() {
		time.Sleep(sleepTime)
		m.run(t)
	}()
}

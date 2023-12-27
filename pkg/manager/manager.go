package manager

import (
	"net/http"
	"time"

	"github.com/SergeyCherepiuk/fleet/pkg/collections/queue"
	"github.com/SergeyCherepiuk/fleet/pkg/consensus"
	"github.com/SergeyCherepiuk/fleet/pkg/httpclient"
	"github.com/SergeyCherepiuk/fleet/pkg/node"
	"github.com/SergeyCherepiuk/fleet/pkg/scheduler"
	"github.com/SergeyCherepiuk/fleet/pkg/task"
	"github.com/SergeyCherepiuk/fleet/pkg/worker"
	"github.com/google/uuid"
	"golang.org/x/exp/maps"
)

const (
	EventQueueInterval   = 100 * time.Millisecond
	MessageQueueInterval = 100 * time.Millisecond
	HeartbeatInterval    = 2 * time.Second

	RestartSleepTimeCoefficient = 2
)

type Manager struct {
	id            uuid.UUID
	node          node.Node
	scheduler     scheduler.Scheduler
	store         consensus.Store
	eventsQueue   queue.Queue[task.Event]
	messagesQueue queue.Queue[worker.Message]
}

func New(node node.Node, scheduler scheduler.Scheduler) *Manager {
	manager := Manager{
		id:            uuid.New(),
		node:          node,
		scheduler:     scheduler,
		store:         consensus.NewLocalStore(),
		eventsQueue:   queue.New[task.Event](0),
		messagesQueue: queue.New[worker.Message](0),
	}

	go manager.watchEventsQueue()
	go manager.watchMessagesQueue()
	go manager.sendHeartbeats()

	return &manager
}

func (m *Manager) AddWorker(wid uuid.UUID, addr node.Addr) {
	cmd := consensus.SetWorkerCommand{
		Index:    m.store.LastIndex() + 1,
		WorkerId: wid,
		Worker: consensus.Worker{
			Addr:  addr,
			Tasks: make(map[uuid.UUID]task.Task),
		},
	}
	m.store.CommitChange(cmd) // Error is ignored (SetWorker command cannot return an error)

	// TODO(SergeyCherepiuk): Broadcast the command
}

func (m *Manager) RemoveWorker(wid uuid.UUID) error {
	cmd := consensus.RemoveWorkerCommand{
		Index:    m.store.LastIndex() + 1,
		WorkerId: wid,
	}
	if _, err := m.store.CommitChange(cmd); err != nil {
		return err
	}

	// TODO(SergeyCherepiuk): Broadcast the command
	return nil
}

func (m *Manager) Run(t task.Task) {
	e := task.Event{Task: t, Desired: task.Running}
	m.eventsQueue.Enqueue(e)
}

func (m *Manager) Stop(tid uuid.UUID) error {
	t, err := m.store.GetTask(tid)
	if err != nil {
		return err
	}

	e := task.Event{Task: t, Desired: task.Finished}
	m.eventsQueue.Enqueue(e)
	return nil
}

func (m *Manager) Tasks() map[uuid.UUID][]task.Task {
	stat := make(map[uuid.UUID][]task.Task)
	for id, workerEntry := range m.store.AllWorkers() {
		stat[id] = maps.Values(workerEntry.Tasks)
	}
	return stat
}

func (m *Manager) watchEventsQueue() {
	for {
		if m.store.Size() == 0 {
			continue
		}

		event, err := m.eventsQueue.Dequeue()
		if err != nil {
			time.Sleep(EventQueueInterval)
			continue
		}

		switch event.Desired {
		case task.Running:
			m.run(event.Task)
		case task.Finished:
			m.finish(event.Task)
		case task.Restarting:
			// TODO(SergeyCherepiuk): Disregard number of restarts if task is
			// running successfully long enough
			m.scheduleRestart(event.Task)
		}
	}
}

func (m *Manager) watchMessagesQueue() {
	for {
		message, err := m.messagesQueue.Dequeue()
		if err != nil {
			time.Sleep(MessageQueueInterval)
			continue
		}

		switch message.Task.State {
		case task.Running, task.Finished, task.Failed:
			cmd := consensus.SetTaskCommand{
				Index:    m.store.LastIndex() + 1,
				WorkerId: message.From,
				Task:     message.Task,
			}
			m.store.CommitChange(cmd) // TODO(SergeyCherepiuk): Handle the error

			// TODO(SergeyCherepiuk): Broadcast the command
		}

		if message.Task.State == task.Restarting {
			event := task.Event{Task: message.Task, Desired: task.Restarting}
			m.eventsQueue.Enqueue(event)
		}
	}
}

func (m *Manager) sendHeartbeats() {
	for {
		for wid, worker := range m.store.AllWorkers() {
			resp, err := httpclient.Get(worker.Addr.String(), "/heartbeat")
			if err != nil || resp.StatusCode != http.StatusOK {
				cmd := consensus.RemoveWorkerCommand{
					Index:    m.store.LastIndex() + 1,
					WorkerId: wid,
				}
				m.store.CommitChange(cmd) // TODO(SergeyCherepiuk): Handle the error

				// TODO(SergeyCherepiuk): Broadcast the command

				for _, t := range worker.Tasks {
					m.Run(t)
				}
			}
		}
		time.Sleep(HeartbeatInterval)
	}
}

func (m *Manager) run(t task.Task) error {
	workers := m.store.AllWorkers()
	workerId, worker, err := m.scheduler.SelectWorker(t, workers)
	if err != nil {
		return err
	}

	t.State = task.Scheduled

	cmd := consensus.SetTaskCommand{
		Index:    m.store.LastIndex() + 1,
		WorkerId: workerId,
		Task:     t,
	}
	m.store.CommitChange(cmd) // TODO(SergeyCherepiuk): Handle the error

	// TODO(SergeyCherepiuk): Broadcast the command

	httpclient.Post(worker.Addr.String(), "/task/run", t)
	return nil
}

func (m *Manager) finish(t task.Task) error {
	_, worker, err := m.store.GetWorkerByTaskId(t.Id)
	if err != nil {
		return err
	}

	httpclient.Post(worker.Addr.String(), "/task/stop", t)
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
		t.Restarts = append(t.Restarts, time.Now())
		m.run(t)
	}()
}

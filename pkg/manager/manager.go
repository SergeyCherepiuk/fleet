package manager

import (
	"net/http"
	"time"

	httpinternal "github.com/SergeyCherepiuk/fleet/internal/http"
	"github.com/SergeyCherepiuk/fleet/pkg/collections/queue"
	"github.com/SergeyCherepiuk/fleet/pkg/consensus"
	"github.com/SergeyCherepiuk/fleet/pkg/container"
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
	BackOffResetInterval = 5 * time.Minute

	BackOffTimeCoefficient = 2
)

type Manager struct {
	id                  uuid.UUID
	node                node.Node
	scheduler           scheduler.Scheduler
	Store               consensus.Store
	EventsQueue         *queue.TimeBasedQueue[task.Event]
	WorkerMessagesQueue *queue.Queue[worker.Message]
}

func New(node node.Node, scheduler scheduler.Scheduler) *Manager {
	manager := Manager{
		id:                  uuid.New(),
		node:                node,
		scheduler:           scheduler,
		Store:               consensus.NewLocalStore(),
		EventsQueue:         queue.NewTimeBasedQueue[task.Event](0, EventQueueInterval),
		WorkerMessagesQueue: queue.NewQueue[worker.Message](0),
	}

	go manager.watchEventsQueue()
	go manager.watchWorkerMessageQueue()
	go manager.sendHeartbeats()

	return &manager
}

func (m *Manager) AddWorker(wid uuid.UUID, addr node.Addr) {
	worker := consensus.Worker{
		Addr:  addr,
		Tasks: make(map[uuid.UUID]task.Task),
	}
	cmd := consensus.NewSetWorkerCommand(m.Store.LastIndex()+1, wid, worker)
	m.Store.CommitChange(*cmd) // Error is ignored (SetWorker command cannot return an error)
}

func (m *Manager) RemoveWorker(wid uuid.UUID) error {
	cmd := consensus.NewRemoveWorkerCommand(m.Store.LastIndex()+1, wid)
	if _, err := m.Store.CommitChange(*cmd); err != nil {
		return err
	}
	return nil
}

func (m *Manager) WorkerTasks(wid uuid.UUID) []task.Task {
	w, err := m.Store.GetWorker(wid)
	if err != nil {
		return make([]task.Task, 0)
	}
	return maps.Values(w.Tasks)
}

func (m *Manager) Tasks() []task.Task {
	tasks := make([]task.Task, 0)
	for _, worker := range m.Store.AllWorkers() {
		tasks = append(tasks, maps.Values(worker.Tasks)...)
	}
	return tasks
}

func (m *Manager) watchEventsQueue() {
	for event := range m.EventsQueue.Out() {
		if m.Store.WorkersNumber() == 0 { // NOTE(SergeyCherepiuk): No workers available
			m.EventsQueue.Enqueue(event)
			time.Sleep(EventQueueInterval)
			continue
		}

		var err error
		switch event.Desired {
		case task.Running:
			err = m.run(event.Task)
		case task.Finished:
			err = m.finish(event.Task)
		case task.RestartingImmediately:
			m.restart(event.Task)
		case task.RestartingWithBackOff:
			m.scheduleRestart(event.Task)
		}

		if err != nil {
			m.EventsQueue.Enqueue(event)
			time.Sleep(EventQueueInterval)
			continue
		}
	}
}

func (m *Manager) watchWorkerMessageQueue() {
	for {
		message, err := m.WorkerMessagesQueue.Dequeue()
		if err != nil {
			time.Sleep(MessageQueueInterval)
			continue
		}

		t := message.Task

		lastIndex := m.Store.LastIndex()
		cmd := consensus.NewSetTaskCommand(lastIndex+1, message.From, t)
		m.Store.CommitChange(*cmd)

		if t.State.Fail() || t.State == task.Finished {
			rp := message.Task.Container.Config.RestartPolicy

			shouldBeRestarted := rp == container.Always ||
				(rp == container.OnFailure && t.State.Fail())

			if shouldBeRestarted {
				var restartMethod task.State
				if t.State == task.FailedOnStartup {
					restartMethod = task.RestartingWithBackOff
				} else {
					restartMethod = task.RestartingImmediately
				}

				cmd := consensus.NewRemoveTaskCommand(lastIndex+2, t.Id)
				m.Store.CommitChange(*cmd)

				event := task.Event{Task: t, Desired: restartMethod}
				m.EventsQueue.Enqueue(event)
			}
		}
	}
}

func (m *Manager) sendHeartbeats() {
	for range time.NewTicker(HeartbeatInterval).C {
		for wid, worker := range m.Store.AllWorkers() {
			resp, err := httpclient.Post(worker.Addr.String(), "/heartbeat", m.Store.LastIndex())

			rescheduleTasks := err != nil || resp == nil ||
				resp.Body == nil || resp.StatusCode != http.StatusOK

			if rescheduleTasks {
				cmd := consensus.NewRemoveWorkerCommand(m.Store.LastIndex()+1, wid)
				m.Store.CommitChange(*cmd)

				for _, t := range worker.Tasks {
					t.State = task.FailedAfterStartup
					event := task.Event{Task: t, Desired: task.RestartingImmediately}
					m.EventsQueue.Enqueue(event)
				}

				continue
			}

			var off int
			if err := httpinternal.Body(resp, &off); err != nil {
				continue
			}

			if off > 0 {
				cmds := m.Store.GetLastNCommands(off)
				go m.broadcastCommandsToWorker(worker.Addr, cmds...)
			}
		}
	}
}

func (m *Manager) run(t task.Task) error {
	workers := m.Store.AllWorkers()
	workerId, worker, err := m.scheduler.SelectWorker(t, workers)
	if err != nil {
		return err
	}

	t.State = task.Scheduled

	cmd := consensus.NewSetTaskCommand(m.Store.LastIndex()+1, workerId, t)
	m.Store.CommitChange(*cmd)

	httpclient.Post(worker.Addr.String(), "/task/run", t)
	return nil
}

func (m *Manager) finish(t task.Task) error {
	_, worker, err := m.Store.GetWorkerByTaskId(t.Id)
	if err != nil {
		return err
	}

	httpclient.Post(worker.Addr.String(), "/task/stop", t)
	return nil
}

func (m *Manager) restart(t task.Task) {
	t.Restarts = append(t.Restarts, time.Now())
	event := task.Event{Task: t, Desired: task.Running}
	m.EventsQueue.Enqueue(event)
}

func (m *Manager) scheduleRestart(t task.Task) {
	var backOffTime time.Duration

	restartWithoutBackOff := len(t.Restarts) == 0 ||
		t.FinishedAt.Sub(t.StartedAt) > BackOffResetInterval

	if restartWithoutBackOff {
		backOffTime = 1 * time.Second
	} else {
		lastBackOffTime := time.Since(t.Restarts[len(t.Restarts)-1])
		backOffTime = lastBackOffTime * BackOffTimeCoefficient
	}

	t.Restarts = append(t.Restarts, time.Now())
	processAfter := time.Now().Add(backOffTime)
	event := task.Event{Task: t, Desired: task.Running}
	m.EventsQueue.EnqueueWithDelay(processAfter, event)
}

func (m *Manager) broadcastCommandsToWorker(addr node.Addr, cmds ...consensus.Command) {
	resp, err := httpclient.Post(addr.String(), "/store/command", cmds)
	if err != nil {
		return
	}

	if resp.StatusCode == http.StatusCreated {
		return
	}

	var off int
	httpinternal.Body(resp, &off)

	if off > 0 {
		cmds = m.Store.GetLastNCommands(off)
		m.broadcastCommandsToWorker(addr, cmds...)
	}
}

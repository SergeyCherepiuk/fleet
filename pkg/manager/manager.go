package manager

import (
	"github.com/SergeyCherepiuk/fleet/pkg/scheduler"
	"github.com/SergeyCherepiuk/fleet/pkg/task"
	"github.com/SergeyCherepiuk/fleet/pkg/worker"
)

type Manager struct {
	Scheduler scheduler.Scheduler
	Workers   []worker.Worker // TODO: change to []uuid.UUID
	Pending   chan task.Task
}

func (m Manager) Run() {
	for {
		select {
		case t := <-m.Pending:
			worker, err := m.Scheduler.SelectWorker(t, m.Workers)
			if err != nil {
				continue
			}

			event := task.NewEvent(t, task.Running)
			sendEvent(worker, event)
		}
	}
}

func sendEvent(worker worker.Worker, event task.Event) error {
	return worker.Execute(event)
}

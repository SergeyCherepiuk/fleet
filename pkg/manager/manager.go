package manager

import (
	"time"

	"github.com/SergeyCherepiuk/fleet/pkg/collections/queue"
	"github.com/SergeyCherepiuk/fleet/pkg/node"
	"github.com/SergeyCherepiuk/fleet/pkg/scheduler"
	"github.com/SergeyCherepiuk/fleet/pkg/task"
	"github.com/google/uuid"
)

type Manager struct {
	node.Node
	ID           uuid.UUID
	Scheduler    scheduler.Scheduler
	WorkerNodes  []node.Addr
	PendingTasks queue.Queue[task.Task]
}

func (m *Manager) Run() {
	for {
		if m.PendingTasks.IsEmpty() {
			time.Sleep(100 * time.Millisecond)
			continue
		}

		// t, err := m.PendingTasks.Dequeue()
		// if err != nil {
		// 	continue
		// }
		//
		// w, err := m.Scheduler.SelectWorker(t, m.WorkerNodes)
		// if err != nil {
		// 	continue
		// }
		//
		// e := task.NewEvent(t, task.Running)
		// w.Execute(e)
	}
}

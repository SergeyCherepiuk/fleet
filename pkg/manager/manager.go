package manager

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/SergeyCherepiuk/fleet/pkg/collections/queue"
	"github.com/SergeyCherepiuk/fleet/pkg/node"
	"github.com/SergeyCherepiuk/fleet/pkg/scheduler"
	"github.com/SergeyCherepiuk/fleet/pkg/task"
	"github.com/SergeyCherepiuk/fleet/pkg/worker"
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
	// TODO(SergeyCherepiuk): Explore if observer pattern applicable
	for {
		if m.PendingTasks.IsEmpty() {
			time.Sleep(100 * time.Millisecond)
			continue
		}

		t, err := m.PendingTasks.Dequeue()
		if err != nil {
			continue
		}

		w, err := m.Scheduler.SelectWorker(t, m.WorkerNodes)
		if err != nil {
			continue
		}

		marshaledTask, err := json.Marshal(t)
		if err != nil {
			continue
		}

		workerAddr := fmt.Sprintf("%s:%d", w.Addr, w.Port)
		url, err := url.JoinPath("http://", workerAddr, worker.TaskRunEndpoint)
		body := bytes.NewReader(marshaledTask)

		// TODO(SergeyCherepiuk): Process response
		if _, err := http.Post(url, "application/json", body); err != nil {
			m.PendingTasks.Enqueue(t) // TODO(SergeyCherepiuk): Re-enqueue up to N times
		}
	}
}

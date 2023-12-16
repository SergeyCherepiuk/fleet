package manager

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

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
	WorkersAddrs []node.Addr
}

func (m *Manager) Run(t task.Task) error {
	w, err := m.Scheduler.SelectWorker(t, m.WorkersAddrs)
	if err != nil {
		return err
	}
	t.State = task.Scheduled

	marshaledTask, err := json.Marshal(t)
	if err != nil {
		return err
	}

	workerAddr := fmt.Sprintf("%s:%d", w.Addr, w.Port)
	url, err := url.JoinPath("http://", workerAddr, worker.TaskRunEndpoint)
	body := bytes.NewReader(marshaledTask)

	// TODO(SergeyCherepiuk): Process response
	_, err = http.Post(url, "application/json", body)
	return err
}

func (m *Manager) Finish(taskId uuid.UUID) error {
	// TODO(SergeyCherepiuk): Find which worker is running the task
	w := node.Addr{}

	workerAddr := fmt.Sprintf("%s:%d", w.Addr, w.Port)
	url, err := url.JoinPath("http://", workerAddr, worker.TaskFinishEndpoint, taskId.String())

	// TODO(SergeyCherepiuk): Process response
	_, err = http.Post(url, "", nil)
	return err
}

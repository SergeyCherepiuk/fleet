package task

import (
	"time"

	"github.com/SergeyCherepiuk/fleet/pkg/container"
	"github.com/google/uuid"
)

type State string

const (
	Pending    State = "pending"
	Scheduled  State = "scheduled"
	Running    State = "running"
	Finished   State = "finished"
	Failed     State = "failed"
	Restarting State = "restarting"
)

type Task struct {
	ID        uuid.UUID
	State     State
	Container container.Container
	Restarts  []time.Time

	StartedAt  time.Time
	FinishedAt time.Time
}

func New(container container.Container) *Task {
	return &Task{
		ID:        uuid.New(),
		State:     Pending,
		Container: container,
		Restarts:  make([]time.Time, 0),
	}
}

type Event struct {
	Task    Task
	Desired State
}

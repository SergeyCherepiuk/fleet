package task

import (
	"time"

	"github.com/SergeyCherepiuk/fleet/pkg/container"
	"github.com/google/uuid"
)

type State string

const (
	Pending   State = "pending"
	Scheduled State = "scheduled"
	Running   State = "running"
	Finished  State = "finished"
	Failed    State = "failed"
)

type Task struct {
	ID        uuid.UUID
	State     State
	Container container.Container

	StartedAt  time.Time
	FinishedAt time.Time
}

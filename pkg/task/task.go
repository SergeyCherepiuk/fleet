package task

import (
	"time"

	"github.com/SergeyCherepiuk/fleet/pkg/container"
	"github.com/google/uuid"
)

type Task struct {
	ID        uuid.UUID
	State     State
	Container container.Container

	StartedAt  time.Time
	FinishedAt time.Time
}

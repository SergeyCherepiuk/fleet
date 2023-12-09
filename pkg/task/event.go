package task

import (
	"time"

	"github.com/google/uuid"
)

type Event struct {
	ID          uuid.UUID
	Task        Task
	State       State
	RequestedAt time.Time
}

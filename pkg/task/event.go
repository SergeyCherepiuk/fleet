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

func NewEvent(task Task, state State) Event {
	return Event{
		ID:          uuid.New(),
		Task:        task,
		State:       state,
		RequestedAt: time.Now(),
	}
}

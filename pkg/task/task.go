package task

import (
	"time"

	"github.com/SergeyCherepiuk/fleet/pkg/container"
	"github.com/google/uuid"
)

type State string

func (s State) Fail() bool {
	return s == FailedOnStartup || s == FailedAfterStartup
}

const (
	Pending               State = "Pending"
	Scheduled             State = "Scheduled"
	Running               State = "Running"
	Finished              State = "Finished"
	FailedOnStartup       State = "FailedOnStartup"
	FailedAfterStartup    State = "FailedAfterStartup"
	RestartingWithBackOff State = "RestartingWithBackOff"
)

type Task struct {
	Id        uuid.UUID
	State     State
	Container container.Container

	StartedAt  []time.Time
	FinishedAt []time.Time
}

func New(container container.Container) *Task {
	return &Task{
		Id:         uuid.New(),
		State:      Pending,
		Container:  container,
		StartedAt:  make([]time.Time, 0),
		FinishedAt: make([]time.Time, 0),
	}
}

type Event struct {
	Task    Task
	Desired State
}

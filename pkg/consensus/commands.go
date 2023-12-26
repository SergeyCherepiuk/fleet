package consensus

import (
	"github.com/SergeyCherepiuk/fleet/pkg/task"
	"github.com/google/uuid"
)

type Command interface {
	GetIndex() int
}

type SetWorkerCommand struct {
	Index    int
	WorkerId uuid.UUID
	Worker   Worker
}

func (c SetWorkerCommand) GetIndex() int {
	return c.Index
}

type RemoveWorkerCommand struct {
	Index    int
	WorkerId uuid.UUID
}

func (c RemoveWorkerCommand) GetIndex() int {
	return c.Index
}

type SetTaskCommand struct {
	Index    int
	WorkerId uuid.UUID
	Task     task.Task
}

func (c SetTaskCommand) GetIndex() int {
	return c.Index
}

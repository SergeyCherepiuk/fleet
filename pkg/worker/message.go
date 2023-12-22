package worker

import (
	"github.com/SergeyCherepiuk/fleet/pkg/task"
	"github.com/google/uuid"
)

type Message struct {
	From uuid.UUID
	Task task.Task
}

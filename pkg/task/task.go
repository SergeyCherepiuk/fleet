package task

import (
	"github.com/SergeyCherepiuk/fleet/pkg/container"
	"github.com/google/uuid"
)

type Task struct {
	ID        uuid.UUID
	Name      string
	State     State
	Container container.Container
}

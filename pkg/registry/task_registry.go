package registry

import (
	"fmt"

	"github.com/SergeyCherepiuk/fleet/pkg/task"
	"github.com/google/uuid"
	"golang.org/x/exp/maps"
)

type TaskRegistry map[uuid.UUID]task.Task

func (tr *TaskRegistry) Get(tid uuid.UUID) (task.Task, error) {
	if t, ok := (*tr)[tid]; ok {
		return t, nil
	}
	return task.Task{}, taskNotFound(tid)
}

func (tr *TaskRegistry) GetAll() []task.Task {
	return maps.Values(*tr)
}

func (tr *TaskRegistry) Set(t task.Task) {
	(*tr)[t.ID] = t
}

func (tr *TaskRegistry) Remove(tid uuid.UUID) error {
	if _, ok := (*tr)[tid]; !ok {
		return taskNotFound(tid)
	}

	delete(*tr, tid)
	return nil
}

type ErrTaskNotFound error

func taskNotFound(tid uuid.UUID) error {
	return ErrTaskNotFound(fmt.Errorf("task with id=%q not found", tid))
}

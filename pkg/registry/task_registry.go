package registry

import (
	"errors"

	"github.com/SergeyCherepiuk/fleet/pkg/task"
	"github.com/google/uuid"
	"golang.org/x/exp/maps"
)

var ErrTaskNotFound = errors.New("task is not found")

type TaskRegistry map[uuid.UUID]task.Task

func (tr *TaskRegistry) Get(tid uuid.UUID) (task.Task, error) {
	if t, ok := (*tr)[tid]; ok {
		return t, nil
	}
	return task.Task{}, ErrTaskNotFound
}

func (tr *TaskRegistry) GetAll() []task.Task {
	return maps.Values(*tr)
}

func (tr *TaskRegistry) Set(t task.Task) {
	(*tr)[t.Id] = t
}

func (tr *TaskRegistry) Remove(tid uuid.UUID) error {
	if _, ok := (*tr)[tid]; !ok {
		return ErrTaskNotFound
	}

	delete(*tr, tid)
	return nil
}

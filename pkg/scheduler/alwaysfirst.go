package scheduler

import (
	"errors"

	"github.com/SergeyCherepiuk/fleet/pkg/manager/registry"
	"github.com/SergeyCherepiuk/fleet/pkg/task"
	"github.com/google/uuid"
	"golang.org/x/exp/maps"
)

type AlwaysFirst struct{}

func (s AlwaysFirst) SelectWorker(task task.Task, workers map[uuid.UUID]registry.Entry) (uuid.UUID, registry.Entry, error) {
	keys := maps.Keys(workers)
	if len(keys) == 0 {
		return uuid.Nil, registry.Entry{}, ErrNoWorkersAvailable(errors.New("no workers available"))
	}
	return keys[0], workers[keys[0]], nil
}

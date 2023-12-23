package scheduler

import (
	"sort"

	"github.com/SergeyCherepiuk/fleet/pkg/registry"
	"github.com/SergeyCherepiuk/fleet/pkg/task"
	"github.com/google/uuid"
	"golang.org/x/exp/maps"
)

type RoundRobin struct {
	last int
}

func (s *RoundRobin) SelectWorker(t task.Task, wes map[uuid.UUID]registry.WorkerEntry) (uuid.UUID, registry.WorkerEntry, error) {
	if len(wes) == 0 {
		return uuid.Nil, registry.WorkerEntry{}, ErrNoWorkersAvailable
	}

	if s.last+1 < len(wes) {
		s.last++
	} else {
		s.last = 0
	}

	keys := maps.Keys(wes)
	sort.Slice(keys, func(i, j int) bool {
		return keys[i].String() < keys[j].String()
	})

	return keys[s.last], wes[keys[s.last]], nil
}

package scheduler

import (
	"sort"

	"github.com/SergeyCherepiuk/fleet/pkg/consensus"
	"github.com/SergeyCherepiuk/fleet/pkg/task"
	"github.com/google/uuid"
	"golang.org/x/exp/maps"
)

type AlwaysFirst struct{}

func (s *AlwaysFirst) SelectWorker(t task.Task, wes map[uuid.UUID]consensus.Worker) (uuid.UUID, consensus.Worker, error) {
	if len(wes) > 0 {
		keys := maps.Keys(wes)
		sort.Slice(keys, func(i, j int) bool {
			return keys[i].String() < keys[j].String()
		})
		return keys[0], wes[keys[0]], nil
	}

	return uuid.Nil, consensus.Worker{}, ErrNoWorkersAvailable
}

package scheduler

import (
	"sort"

	"github.com/SergeyCherepiuk/fleet/pkg/consensus"
	"github.com/SergeyCherepiuk/fleet/pkg/task"
	"github.com/google/uuid"
	"golang.org/x/exp/maps"
)

type roundRobin struct {
	last int
}

func NewRoundRobin() *roundRobin {
	return &roundRobin{last: 0}
}

func (s *roundRobin) SelectWorker(t task.Task, ws map[uuid.UUID]consensus.Worker) (uuid.UUID, consensus.Worker, error) {
	if len(ws) == 0 {
		return uuid.Nil, consensus.Worker{}, ErrNoAvailableWorkers
	}

	if s.last+1 < len(ws) {
		s.last++
	} else {
		s.last = 0
	}

	keys := maps.Keys(ws)
	sort.Slice(keys, func(i, j int) bool {
		return keys[i].String() < keys[j].String()
	})

	return keys[s.last], ws[keys[s.last]], nil
}

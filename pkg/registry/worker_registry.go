package registry

import (
	"fmt"
	"time"

	"github.com/SergeyCherepiuk/fleet/pkg/node"
	"github.com/google/uuid"
	"golang.org/x/exp/maps"
)

const Lifetime = 30 * time.Second

type WorkerRegistry map[uuid.UUID]WorkerEntry

type WorkerEntry struct {
	ID    uuid.UUID
	Addr  node.Addr
	Tasks TaskRegistry
	exp   time.Time
}

func (wr *WorkerRegistry) Watch() {
	for {
		for id, we := range *wr {
			if we.exp.Before(time.Now()) {
				delete(*wr, id)
			}
		}
		time.Sleep(time.Second)
	}
}

func (wr *WorkerRegistry) Renew(wid uuid.UUID) error {
	we, err := wr.Get(wid)
	if err != nil {
		return err
	}

	we.exp = time.Now().Add(Lifetime)
	(*wr)[wid] = we
	return nil
}

func (wr *WorkerRegistry) GetByTaskId(tid uuid.UUID) (WorkerEntry, error) {
	for _, entry := range *wr {
		if _, ok := entry.Tasks[tid]; ok {
			return entry, nil
		}
	}

	return WorkerEntry{}, ErrWorkerNotFound(
		fmt.Errorf("worker that have task with id=%q not found", tid),
	)
}

func (wr *WorkerRegistry) Get(wid uuid.UUID) (WorkerEntry, error) {
	if w, ok := (*wr)[wid]; ok {
		return w, nil
	}
	return WorkerEntry{}, workerNotFound(wid)
}

func (wr *WorkerRegistry) GetAll() []WorkerEntry {
	return maps.Values(*wr)
}

func (wr *WorkerRegistry) Add(wid uuid.UUID, addr node.Addr) {
	(*wr)[wid] = WorkerEntry{
		ID:    uuid.New(),
		Addr:  addr,
		Tasks: make(TaskRegistry),
		exp:   time.Now().Add(Lifetime),
	}
}

func (wr *WorkerRegistry) Set(wid uuid.UUID, we WorkerEntry) error {
	if _, err := wr.Get(wid); err != nil {
		return nil
	}

	(*wr)[wid] = we
	return nil
}

func (wr *WorkerRegistry) Remove(wid uuid.UUID) {
	delete(*wr, wid)
}

type ErrWorkerNotFound error

func workerNotFound(wid uuid.UUID) error {
	return ErrWorkerNotFound(fmt.Errorf("with id=%q not found", wid))
}

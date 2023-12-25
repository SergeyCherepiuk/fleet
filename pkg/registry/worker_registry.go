package registry

import (
	"errors"
	"time"

	"github.com/SergeyCherepiuk/fleet/pkg/node"
	"github.com/SergeyCherepiuk/fleet/pkg/task"
	"github.com/google/uuid"
)

const Lifetime = 30 * time.Second

var ErrWorkerNotFound = errors.New("worker is not found")

type WorkerRegistry map[uuid.UUID]WorkerEntry

type WorkerEntry struct {
	Addr  node.Addr
	Tasks TaskRegistry
}

func NewWorkerRegistry() WorkerRegistry {
	wr := make(WorkerRegistry)
	return wr
}

func NewWorkerEntry(addr node.Addr) WorkerEntry {
	return WorkerEntry{
		Addr:  addr,
		Tasks: make(TaskRegistry),
	}
}

func (wr *WorkerRegistry) GetByTaskId(tid uuid.UUID) (uuid.UUID, WorkerEntry, error) {
	for id, entry := range *wr {
		if _, ok := entry.Tasks[tid]; ok {
			return id, entry, nil
		}
	}

	return uuid.Nil, WorkerEntry{}, ErrWorkerNotFound
}

func (wr *WorkerRegistry) Get(wid uuid.UUID) (WorkerEntry, error) {
	if w, ok := (*wr)[wid]; ok {
		return w, nil
	}
	return WorkerEntry{}, ErrWorkerNotFound
}

func (wr *WorkerRegistry) GetAll() map[uuid.UUID]WorkerEntry {
	return *wr
}

func (wr *WorkerRegistry) Set(wid uuid.UUID, we WorkerEntry) {
	(*wr)[wid] = we
}

func (wr *WorkerRegistry) SetTask(wid uuid.UUID, t task.Task) error {
	workerEntry, err := wr.Get(wid)
	if err != nil {
		return err
	}

	workerEntry.Tasks.Set(t)
	wr.Set(wid, workerEntry)
	return nil
}

func (wr *WorkerRegistry) Remove(wid uuid.UUID) (TaskRegistry, error) {
	we, ok := (*wr)[wid]
	if !ok {
		return make(TaskRegistry, 0), ErrWorkerNotFound
	}

	delete(*wr, wid)
	return we.Tasks, nil
}

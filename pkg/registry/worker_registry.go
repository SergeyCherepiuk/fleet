package registry

import (
	"errors"
	"time"

	"github.com/SergeyCherepiuk/fleet/pkg/node"
	"github.com/SergeyCherepiuk/fleet/pkg/task"
	"github.com/google/uuid"
	"golang.org/x/exp/maps"
)

const Lifetime = 30 * time.Second

var ErrWorkerNotFound = errors.New("worker is not found")

type WorkerRegistry map[uuid.UUID]WorkerEntry

type WorkerEntry struct {
	Addr  node.Addr
	Tasks TaskRegistry
	exp   time.Time
}

func NewWorkerRegistry() WorkerRegistry {
	wr := make(WorkerRegistry)
	go wr.watch()
	return wr
}

func NewWorkerEntry(addr node.Addr) WorkerEntry {
	return WorkerEntry{
		Addr:  addr,
		Tasks: make(TaskRegistry),
		exp:   time.Now().Add(Lifetime),
	}
}

func (wr *WorkerRegistry) watch() {
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

	return WorkerEntry{}, ErrWorkerNotFound
}

func (wr *WorkerRegistry) Get(wid uuid.UUID) (WorkerEntry, error) {
	if w, ok := (*wr)[wid]; ok {
		return w, nil
	}
	return WorkerEntry{}, ErrWorkerNotFound
}

func (wr *WorkerRegistry) GetAll() []WorkerEntry {
	return maps.Values(*wr)
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

func (wr *WorkerRegistry) Remove(wid uuid.UUID) {
	delete(*wr, wid)
}

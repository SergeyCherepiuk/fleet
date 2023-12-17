package registry

import (
	"fmt"
	"time"

	"github.com/SergeyCherepiuk/fleet/pkg/node"
	"github.com/SergeyCherepiuk/fleet/pkg/task"
	"github.com/google/uuid"
)

const Lifetime = 30 * time.Second

type WorkerRegistry struct {
	registry map[uuid.UUID]Entry
}

type Entry struct {
	Addr  node.Addr
	exp   time.Time
	tasks map[uuid.UUID]task.Task
}

type ErrWorkerNotFound error

func New() *WorkerRegistry {
	return &WorkerRegistry{
		registry: make(map[uuid.UUID]Entry),
	}
}

func (wr *WorkerRegistry) Watch() {
	for {
		for id, entry := range wr.registry {
			if entry.exp.Before(time.Now()) {
				delete(wr.registry, id)
			}
		}
		time.Sleep(time.Second)
	}
}

func (wr *WorkerRegistry) Renew(workerId uuid.UUID) error {
	entry, ok := wr.registry[workerId]
	if !ok {
		return ErrWorkerNotFound(fmt.Errorf("worker with id=%q not found", workerId))
	}

	entry.exp = time.Now().Add(Lifetime)
	wr.registry[workerId] = entry
	return nil
}

func (wr *WorkerRegistry) Get(workerId uuid.UUID) (Entry, error) {
	e, ok := wr.registry[workerId]
	if !ok {
		return Entry{}, ErrWorkerNotFound(
			fmt.Errorf("worker with id=%q not found", workerId),
		)
	}
	return e, nil
}

func (wr *WorkerRegistry) GetAll() map[uuid.UUID]Entry {
	return wr.registry
}

func (wr *WorkerRegistry) AddWorker(workerId uuid.UUID, workerAddr node.Addr) {
	wr.registry[workerId] = Entry{
		Addr:  workerAddr,
		exp:   time.Now().Add(Lifetime),
		tasks: make(map[uuid.UUID]task.Task),
	}
}

func (wr *WorkerRegistry) RemoveWorker(workerId uuid.UUID) {
	delete(wr.registry, workerId)
}

func (wr *WorkerRegistry) FindWorker(taskId uuid.UUID) (uuid.UUID, Entry, error) {
	for id, entry := range wr.registry {
		if _, ok := entry.tasks[taskId]; ok {
			return id, entry, nil
		}
	}

	return uuid.Nil, Entry{}, ErrWorkerNotFound(
		fmt.Errorf("worker that have task with id=%q not found", taskId),
	)
}

func (wr *WorkerRegistry) AddTask(workerId uuid.UUID, task task.Task) error {
	entry, err := wr.Get(workerId)
	if err != nil {
		return err
	}

	entry.tasks[task.ID] = task
	wr.registry[workerId] = entry
	return nil
}

func (wr *WorkerRegistry) RemoveTask(taskId uuid.UUID) error {
	_, entry, err := wr.FindWorker(taskId)
	if err != nil {
		return err
	}

	delete(entry.tasks, taskId) // TODO(SergeyCherepiuk): Need to re-set the entry?
	return nil
}

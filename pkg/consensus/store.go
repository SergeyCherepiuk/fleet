package consensus

import (
	"encoding/json"
	"errors"
	"sync"

	mapsinternal "github.com/SergeyCherepiuk/fleet/internal/maps"
	"github.com/SergeyCherepiuk/fleet/pkg/node"
	"github.com/SergeyCherepiuk/fleet/pkg/task"
	"github.com/google/uuid"
)

type Store interface {
	AllWorkers() map[uuid.UUID]Worker
	GetTask(taskId uuid.UUID) (task.Task, error)
	GetWorker(worketId uuid.UUID) (Worker, error)
	GetWorkerByTaskId(taskId uuid.UUID) (uuid.UUID, Worker, error)
	GetLastNCommands(n int) []Command
	Size() int

	LastIndex() int
	CommitChange(cmd Command) (off int, err error)
}

var (
	ErrLogOutOfSync   = errors.New("log is out of sync")
	ErrWorkerNotFound = errors.New("worker is not found")
	ErrTaskNotFound   = errors.New("task is not found")
	ErrUnknownCommand = errors.New("unknown command")
)

type store struct {
	muState sync.RWMutex
	state   map[uuid.UUID]Worker

	muLog sync.RWMutex
	log   []Command
}

func NewLocalStore() *store {
	return &store{
		state: make(map[uuid.UUID]Worker),
		log:   make([]Command, 0),
	}
}

type Worker struct {
	Addr  node.Addr
	Tasks map[uuid.UUID]task.Task
}

func (s *store) AllWorkers() map[uuid.UUID]Worker {
	s.muState.RLock()
	defer s.muState.RUnlock()

	c := make(map[uuid.UUID]Worker, len(s.state))
	for k, v := range s.state {
		c[k] = v
	}
	return c
}

func (s *store) GetTask(tid uuid.UUID) (task.Task, error) {
	s.muState.RLock()
	defer s.muState.RUnlock()

	for _, worker := range s.state {
		if task, ok := worker.Tasks[tid]; ok {
			return task, nil
		}
	}
	return task.Task{}, ErrTaskNotFound
}

func (s *store) GetWorker(wid uuid.UUID) (Worker, error) {
	s.muState.RLock()
	defer s.muState.RUnlock()

	if worker, ok := s.state[wid]; ok {
		return worker, nil
	}
	return Worker{}, ErrWorkerNotFound
}

func (s *store) GetWorkerByTaskId(tid uuid.UUID) (uuid.UUID, Worker, error) {
	s.muState.RLock()
	defer s.muState.RUnlock()

	for id, worker := range s.state {
		if _, ok := worker.Tasks[tid]; ok {
			return id, worker, nil
		}
	}
	return uuid.Nil, Worker{}, ErrWorkerNotFound
}

// TODO(SergeyCherepiuk): Called too many times
func (s *store) GetLastNCommands(n int) []Command {
	s.muLog.RLock()
	defer s.muLog.RUnlock()

	n = min(n, len(s.log))
	c := make([]Command, n)
	copy(c, s.log[len(s.log)-n:])
	return c
}

func (s *store) Size() int {
	s.muLog.RLock()
	defer s.muLog.RUnlock()
	return len(s.log)
}

func (s *store) LastIndex() int {
	if s.Size() == 0 {
		return 0
	}

	s.muLog.RLock()
	defer s.muLog.RUnlock()
	return s.log[len(s.log)-1].Index
}

func (s *store) CommitChange(cmd Command) (int, error) {
	lastIndex := s.LastIndex()
	diff := cmd.Index - 1 - lastIndex

	if diff > 0 {
		return diff, ErrLogOutOfSync
	}

	var err error
	switch cmd.Type {
	case SetWorker:
		err = s.setWorker(cmd.Data)
	case RemoveWorker:
		err = s.removeWorker(cmd.Data)
	case SetTask:
		err = s.setTask(cmd.Data)
	case RemoveTask:
		err = s.removeTask(cmd.Data)
	default:
		err = ErrUnknownCommand
	}

	if err != nil {
		return 0, err
	}

	s.muLog.Lock()
	defer s.muLog.Unlock()
	s.log = append(s.log, cmd)
	return 0, nil
}

func (s *store) setWorker(data []byte) error {
	var unmarshaled SetWorkerCommandData
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		return err
	}

	s.muState.Lock()
	defer s.muState.Unlock()

	s.state[unmarshaled.WorkerId] = Worker{
		Addr:  unmarshaled.Worker.Addr,
		Tasks: make(map[uuid.UUID]task.Task),
	}
	return nil
}

func (s *store) removeWorker(data []byte) error {
	var unmarshaled RemoveWorkerCommandData
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		return err
	}

	s.muState.Lock()
	defer s.muState.Unlock()

	if _, ok := s.state[unmarshaled.WorkerId]; !ok {
		return ErrWorkerNotFound
	}

	delete(s.state, unmarshaled.WorkerId)
	return nil
}

func (s *store) setTask(data []byte) error {
	var unmarshaled SetTaskCommandData
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		return err
	}

	s.muState.Lock()
	defer s.muState.Unlock()

	worker, ok := s.state[unmarshaled.WorkerId]
	if !ok {
		return ErrWorkerNotFound
	}

	tasks := mapsinternal.ConcurrentCopy(worker.Tasks)
	tasks[unmarshaled.Task.Id] = unmarshaled.Task
	worker.Tasks = tasks
	s.state[unmarshaled.WorkerId] = worker
	return nil
}

func (s *store) removeTask(data []byte) error {
	var unmarshaled RemoveTaskCommandData
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		return err
	}

	id, worker, err := s.GetWorkerByTaskId(unmarshaled.TaskId)
	if err != nil {
		return err
	}

	tasks := mapsinternal.ConcurrentCopy(worker.Tasks)
	delete(tasks, unmarshaled.TaskId)
	worker.Tasks = tasks

	s.muState.Lock()
	defer s.muState.Unlock()
	s.state[id] = worker
	return nil
}

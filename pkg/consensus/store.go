package consensus

import (
	"encoding/json"
	"errors"

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
	state map[uuid.UUID]Worker
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
	return s.state
}

func (s *store) GetTask(tid uuid.UUID) (task.Task, error) {
	for _, worker := range s.state {
		if task, ok := worker.Tasks[tid]; ok {
			return task, nil
		}
	}
	return task.Task{}, ErrTaskNotFound
}

func (s *store) GetWorker(wid uuid.UUID) (Worker, error) {
	if worker, ok := s.state[wid]; ok {
		return worker, nil
	}
	return Worker{}, ErrWorkerNotFound
}

func (s *store) GetWorkerByTaskId(tid uuid.UUID) (uuid.UUID, Worker, error) {
	for id, worker := range s.state {
		if _, ok := worker.Tasks[tid]; ok {
			return id, worker, nil
		}
	}
	return uuid.Nil, Worker{}, ErrWorkerNotFound
}

func (s *store) GetLastNCommands(n int) []Command {
	return s.log[len(s.log)-n:]
}

func (s *store) Size() int {
	return len(s.state)
}

func (s *store) LastIndex() int {
	if s.Size() == 0 {
		return 0
	}
	return s.log[len(s.log)-1].Index
}

func (s *store) CommitChange(cmd Command) (int, error) {
	if s.LastIndex() != cmd.Index-1 {
		return cmd.Index - s.LastIndex(), ErrLogOutOfSync
	}

	var err error
	switch cmd.Type {
	case SetWorker:
		err = s.setWorker(cmd.Data)
	case RemoveWorker:
		err = s.removeWorker(cmd.Data)
	case SetTask:
		err = s.setTask(cmd.Data)
	default:
		err = ErrUnknownCommand
	}

	if err != nil {
		return 0, err
	}

	s.log = append(s.log, cmd)
	return 0, nil
}

func (s *store) setWorker(data []byte) error {
	var unmarshaled SetWorkerCommandData
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		return err
	}

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

	worker, ok := s.state[unmarshaled.WorkerId]
	if !ok {
		return ErrWorkerNotFound
	}

	worker.Tasks[unmarshaled.Task.Id] = unmarshaled.Task
	s.state[unmarshaled.WorkerId] = worker
	return nil
}

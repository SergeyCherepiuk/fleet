package consensus

import (
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

func (s *store) Size() int {
	return len(s.state)
}

func (s *store) LastIndex() int {
	if s.Size() == 0 {
		return -1
	}
	return s.log[len(s.log)-1].GetIndex()
}

func (s *store) CommitChange(cmd Command) (int, error) {
	if s.LastIndex() != cmd.GetIndex()-1 {
		return cmd.GetIndex() - s.LastIndex(), ErrLogOutOfSync
	}

	var err error
	switch cmd := cmd.(type) {
	case SetWorkerCommand:
		s.setWorker(cmd)
	case RemoveWorkerCommand:
		err = s.removeWorker(cmd)
	case SetTaskCommand:
		err = s.setTask(cmd)
	default:
		err = ErrUnknownCommand
	}

	if err != nil {
		return 0, err
	}

	s.log = append(s.log, cmd)
	return 0, nil
}

func (s *store) setWorker(cmd SetWorkerCommand) {
	s.state[cmd.WorkerId] = Worker{
		Addr:  cmd.Worker.Addr,
		Tasks: make(map[uuid.UUID]task.Task),
	}
}

func (s *store) removeWorker(cmd RemoveWorkerCommand) error {
	if _, ok := s.state[cmd.WorkerId]; !ok {
		return ErrWorkerNotFound
	}

	delete(s.state, cmd.WorkerId)
	return nil
}

func (s *store) setTask(cmd SetTaskCommand) error {
	worker, ok := s.state[cmd.WorkerId]
	if !ok {
		return ErrWorkerNotFound
	}

	worker.Tasks[cmd.Task.Id] = cmd.Task
	s.state[cmd.WorkerId] = worker
	return nil
}

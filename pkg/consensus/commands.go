package consensus

import (
	"encoding/json"

	"github.com/SergeyCherepiuk/fleet/pkg/task"
	"github.com/google/uuid"
)

type CommandType string

const (
	SetWorker    CommandType = "SetWorker"
	RemoveWorker CommandType = "RemoveWorker"
	SetTask      CommandType = "SetTask"
	RemoveTask   CommandType = "RemoveTask"
)

type Command struct {
	Index int
	Type  CommandType
	Data  []byte
}

func NewSetWorkerCommand(index int, workerId uuid.UUID, worker Worker) *Command {
	data := SetWorkerCommandData{WorkerId: workerId, Worker: worker}
	marshaled, _ := json.Marshal(data)
	return &Command{Index: index, Type: SetWorker, Data: marshaled}
}

func NewRemoveWorkerCommand(index int, workerId uuid.UUID) *Command {
	data := RemoveWorkerCommandData{WorkerId: workerId}
	marshaled, _ := json.Marshal(data)
	return &Command{Index: index, Type: RemoveWorker, Data: marshaled}
}

func NewSetTaskCommand(index int, workerId uuid.UUID, task task.Task) *Command {
	data := SetTaskCommandData{WorkerId: workerId, Task: task}
	marshaled, _ := json.Marshal(data)
	return &Command{Index: index, Type: SetTask, Data: marshaled}
}

func NewRemoveTaskCommand(index int, taskId uuid.UUID) *Command {
	data := RemoveTaskCommandData{TaskId: taskId}
	marshaled, _ := json.Marshal(data)
	return &Command{Index: index, Type: RemoveTask, Data: marshaled}
}

type SetWorkerCommandData struct {
	WorkerId uuid.UUID
	Worker   Worker
}

type RemoveWorkerCommandData struct {
	WorkerId uuid.UUID
}

type SetTaskCommandData struct {
	WorkerId uuid.UUID
	Task     task.Task
}

type RemoveTaskCommandData struct {
	TaskId uuid.UUID
}

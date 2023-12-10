package main

import (
	"time"

	"github.com/SergeyCherepiuk/fleet/pkg/collections/queue"
	"github.com/SergeyCherepiuk/fleet/pkg/container"
	"github.com/SergeyCherepiuk/fleet/pkg/image"
	"github.com/SergeyCherepiuk/fleet/pkg/manager"
	"github.com/SergeyCherepiuk/fleet/pkg/scheduler"
	"github.com/SergeyCherepiuk/fleet/pkg/task"
	"github.com/SergeyCherepiuk/fleet/pkg/worker"
	"github.com/google/uuid"
)

func main() {
	w := worker.Worker{
		ID:    uuid.New(),
		Tasks: make(map[uuid.UUID]task.Task),
		Queue: queue.New[task.Task](0),
	}

	s := scheduler.AlwaysFirst{}

	m := manager.Manager{
		Scheduler: s,
		Workers:   []worker.Worker{w},
		Pending:   make(chan task.Task),
	}

	nginxImage := image.Image{
		ID:      uuid.NewString(),
		Registy: "docker.io/library",
		Tag:     "nginx",
		Version: "alpine",
	}

	nginxContainer := container.Container{
		ID:            uuid.NewString(),
		Image:         nginxImage,
		ExposedPorts:  map[uint16]uint16{80: 80},
		RestartPolicy: container.OnFailure,
		CPU:           15.4,
		Memory:        3873,
		Disk:          153674396,
		StartedAt:     time.Now(),
	}

	go m.Run()
	m.Pending <- task.Task{
		ID:        uuid.New(),
		Name:      "task-1",
		State:     task.Pending,
		Container: nginxContainer,
	}

	<-make(chan struct{}) // TODO: Quick workaround for testing (blocking forever)
}

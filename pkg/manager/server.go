package manager

import (
	"fmt"
	"net/http"

	"github.com/SergeyCherepiuk/fleet/pkg/node"
	"github.com/SergeyCherepiuk/fleet/pkg/task"
	"github.com/labstack/echo/v4"
)

const (
	WorkerAddEndpoint = "/worker/assign"
	TaskRunEndpoint   = "/task/run"
)

func StartServer(addr string, manager Manager) error {
	e := echo.New()
	e.HideBanner = true

	go manager.Run()

	e.POST(WorkerAddEndpoint, func(c echo.Context) error {
		var addr node.Addr
		if err := c.Bind(&addr); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid worker node address")
		}

		manager.WorkerNodes = append(manager.WorkerNodes, addr)
		return c.NoContent(http.StatusCreated)
	})

	e.POST(TaskRunEndpoint, func(c echo.Context) error {
		var t task.Task
		if err := c.Bind(&t); err != nil {
			return echo.NewHTTPError(
				http.StatusBadRequest,
				fmt.Errorf("invalid task format: %w", err),
			)
		}

		manager.PendingTasks.Enqueue(t)
		return c.NoContent(http.StatusCreated)
	})

	return e.Start(addr)
}

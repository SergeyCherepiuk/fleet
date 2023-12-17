package manager

import (
	"fmt"
	"net/http"

	"github.com/SergeyCherepiuk/fleet/pkg/node"
	"github.com/SergeyCherepiuk/fleet/pkg/task"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

const (
	WorkerEndpoint   = "/worker"
	TaskRunEndpoint  = "/task/run"
	TaskStopEndpoint = "/tast/stop"
)

func StartServer(addr string, manager *Manager) error {
	e := echo.New()
	e.HideBanner = true

	workerEndpointWithId := fmt.Sprintf("%s/:id", WorkerEndpoint)

	e.POST(workerEndpointWithId, func(c echo.Context) error {
		id, err := uuid.Parse(c.Param("id"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid id format")
		}

		var addr node.Addr
		if err := c.Bind(&addr); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid worker node address")
		}

		manager.workerRegistry.AddWorker(id, addr)
		return c.NoContent(http.StatusCreated)
	})

	e.PUT(workerEndpointWithId, func(c echo.Context) error {
		id, err := uuid.Parse(c.Param("id"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid id format")
		}

		if err := manager.workerRegistry.Renew(id); err != nil {
			return echo.NewHTTPError(
				http.StatusInternalServerError,
				fmt.Errorf("failed to renew worker's expiration time %w", err),
			)
		}

		return c.NoContent(http.StatusOK)
	})

	e.DELETE(workerEndpointWithId, func(c echo.Context) error {
		id, err := uuid.Parse(c.Param("id"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid id format")
		}

		var addr node.Addr
		if err := c.Bind(&addr); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid worker node address")
		}

		manager.workerRegistry.RemoveWorker(id)
		return c.NoContent(http.StatusOK)
	})

	e.POST(TaskRunEndpoint, func(c echo.Context) error {
		var t task.Task
		if err := c.Bind(&t); err != nil {
			return echo.NewHTTPError(
				http.StatusBadRequest,
				fmt.Errorf("invalid task format: %w", err),
			)
		}

		if err := manager.Run(t); err != nil {
			return echo.NewHTTPError(
				http.StatusInternalServerError,
				fmt.Errorf("failed to run a task: %w", err),
			)
		}

		return c.NoContent(http.StatusCreated)
	})

	taskStopEndpoint := fmt.Sprintf("%s/:id", TaskStopEndpoint)
	e.POST(taskStopEndpoint, func(c echo.Context) error {
		id, err := uuid.Parse(c.Param("id"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid id format")
		}

		if err := manager.Finish(id); err != nil {
			return echo.NewHTTPError(
				http.StatusInternalServerError,
				fmt.Errorf("failed to stop the task: %w", err),
			)
		}

		return c.NoContent(http.StatusOK)
	})

	return e.Start(addr)
}

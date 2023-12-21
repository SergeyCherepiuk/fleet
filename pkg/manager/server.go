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
	TaskListEndpoint = "/task/list"
)

func StartServer(addr string, manager *Manager) error {
	e := echo.New()
	e.HideBanner = true

	e.GET(TaskListEndpoint, func(c echo.Context) error {
		return c.JSON(http.StatusOK, manager.Tasks())
	})

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

		manager.workerRegistry.Add(id, addr)
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

		manager.workerRegistry.Remove(id)
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

		manager.Run(t)
		return c.NoContent(http.StatusCreated)
	})

	taskStopEndpoint := fmt.Sprintf("%s/:id", TaskStopEndpoint)
	e.POST(taskStopEndpoint, func(c echo.Context) error {
		id, err := uuid.Parse(c.Param("id"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid id format")
		}

		manager.Stop(id)
		return c.NoContent(http.StatusOK)
	})

	return e.Start(addr)
}

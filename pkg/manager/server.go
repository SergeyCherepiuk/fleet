package manager

import (
	"fmt"
	"net/http"
	"reflect"
	"slices"

	"github.com/SergeyCherepiuk/fleet/pkg/node"
	"github.com/SergeyCherepiuk/fleet/pkg/task"
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

	e.POST(WorkerEndpoint, func(c echo.Context) error {
		var addr node.Addr
		if err := c.Bind(&addr); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid worker node address")
		}

		manager.WorkersAddrs = append(manager.WorkersAddrs, addr)
		return c.NoContent(http.StatusCreated)
	})

	e.DELETE(WorkerEndpoint, func(c echo.Context) error {
		var addr node.Addr
		if err := c.Bind(&addr); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid worker node address")
		}

		manager.WorkersAddrs = slices.DeleteFunc(manager.WorkersAddrs, func(a node.Addr) bool {
			return reflect.DeepEqual(a, addr)
		})
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

	taskStopEndpoint := fmt.Sprintf("%s/{id}", TaskStopEndpoint)
	e.POST(taskStopEndpoint, func(c echo.Context) error {
		return nil
	})

	return e.Start(addr)
}

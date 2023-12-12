package worker

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/SergeyCherepiuk/fleet/pkg/task"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

const (
	TaskRunEndpoint    = "/task/run"
	TaskFinishEndpoint = "/task/finish"
)

func StartServer(addr string, worker Worker) error {
	e := echo.New()
	e.HideBanner = true

	e.POST(TaskRunEndpoint, func(c echo.Context) error {
		var t task.Task
		if err := c.Bind(&t); err != nil {
			return echo.NewHTTPError(
				http.StatusBadRequest,
				fmt.Errorf("invalid task format: %w", err),
			)
		}

		if err := worker.Run(t); err != nil {
			return echo.NewHTTPError(
				http.StatusInternalServerError,
				fmt.Errorf("failed to run the task: %w", err),
			)
		}

		return c.NoContent(http.StatusCreated)
	})

	taskFinishEndpoint, err := url.JoinPath(TaskFinishEndpoint, "{id}")
	if err != nil {
		return err
	}

	e.POST(taskFinishEndpoint, func(c echo.Context) error {
		taskId, err := uuid.Parse(c.Param("id"))
		if err != nil {
			return echo.NewHTTPError(
				http.StatusBadRequest,
				fmt.Errorf("invalid task id: %w", err),
			)
		}

		if err := worker.Finish(taskId); err != nil {
			return echo.NewHTTPError(
				http.StatusInternalServerError,
				fmt.Errorf("failed to stop the task: %w", err),
			)
		}

		return c.NoContent(http.StatusOK)
	})

	return e.Start(addr)
}

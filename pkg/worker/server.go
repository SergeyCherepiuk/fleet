package worker

import (
	"context"
	"fmt"
	"net/http"

	"github.com/SergeyCherepiuk/fleet/pkg/task"
	"github.com/labstack/echo/v4"
)

const (
	TaskRunEndpoint    = "/task/run"
	TaskFinishEndpoint = "/task/finish"
)

func StartServer(addr string, worker *Worker) error {
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

		ctx := context.Background()
		if err := worker.Run(ctx, &t); err != nil {
			return echo.NewHTTPError(
				http.StatusInternalServerError,
				fmt.Errorf("failed to run the task: %w", err),
			)
		}

		return c.JSON(http.StatusCreated, t)
	})

	e.POST(TaskFinishEndpoint, func(c echo.Context) error {
		var t task.Task
		if err := c.Bind(&t); err != nil {
			return echo.NewHTTPError(
				http.StatusBadRequest,
				fmt.Errorf("invalid task format: %w", err),
			)
		}

		ctx := context.Background()
		if err := worker.Finish(ctx, &t); err != nil {
			return echo.NewHTTPError(
				http.StatusInternalServerError,
				fmt.Errorf("failed to stop the task: %w", err),
			)
		}

		return c.JSON(http.StatusOK, t)
	})

	return e.Start(addr)
}

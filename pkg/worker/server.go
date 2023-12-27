package worker

import (
	"context"
	"fmt"
	"net/http"

	"github.com/SergeyCherepiuk/fleet/pkg/consensus"
	"github.com/SergeyCherepiuk/fleet/pkg/task"
	"github.com/labstack/echo/v4"
)

func StartServer(addr string, worker *Worker) error {
	e := echo.New()
	e.HideBanner = true

	e.POST("/task/run", func(c echo.Context) error {
		var t task.Task
		if err := c.Bind(&t); err != nil {
			return echo.NewHTTPError(
				http.StatusBadRequest,
				fmt.Errorf("invalid task format: %w", err),
			)
		}

		ctx := context.Background()
		if err := worker.Run(ctx, &t); err != nil {
			c.JSON(http.StatusInternalServerError, t)
		}

		return c.JSON(http.StatusCreated, t)
	})

	e.POST("/task/stop", func(c echo.Context) error {
		var t task.Task
		if err := c.Bind(&t); err != nil {
			return echo.NewHTTPError(
				http.StatusBadRequest,
				fmt.Errorf("invalid task format: %w", err),
			)
		}

		ctx := context.Background()
		if err := worker.Finish(ctx, &t); err != nil {
			return c.JSON(http.StatusInternalServerError, t)
		}

		return c.JSON(http.StatusOK, t)
	})

	e.GET("/stat", func(c echo.Context) error {
		stat, err := worker.node.Resources()
		if err != nil {
			return echo.NewHTTPError(
				http.StatusInternalServerError,
				fmt.Errorf("failed to collect statistics: %w", err),
			)
		}

		return c.JSON(http.StatusOK, stat)
	})

	e.POST("/store/command", func(c echo.Context) error {
		var cmds []consensus.Command
		if err := c.Bind(&cmds); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err)
		}

		if off, err := worker.CommitChanges(cmds...); err != nil {
			return c.JSON(http.StatusInternalServerError, off)
		}

		return c.NoContent(http.StatusCreated)
	})

	e.GET("/heartbeat", func(c echo.Context) error {
		if err := worker.CancleShutdown(); err != nil {
			return c.NoContent(http.StatusInternalServerError)
		}
		return c.NoContent(http.StatusOK)
	})

	return e.Start(addr)
}

package manager

import (
	"fmt"
	"net/http"

	"github.com/SergeyCherepiuk/fleet/pkg/node"
	"github.com/SergeyCherepiuk/fleet/pkg/registry"
	"github.com/SergeyCherepiuk/fleet/pkg/task"
	"github.com/SergeyCherepiuk/fleet/pkg/worker"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func StartServer(addr string, manager *Manager) error {
	e := echo.New()
	e.HideBanner = true

	workerGroup := e.Group("/worker")
	workerWithIDGroup := workerGroup.Group("/:id", parseID)

	workerWithIDGroup.POST("", func(c echo.Context) error {
		var addr node.Addr
		if err := c.Bind(&addr); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid worker node address")
		}

		id := c.Get("id").(uuid.UUID)
		workerEntry := registry.NewWorkerEntry(addr)
		manager.workerRegistry.Set(id, workerEntry)
		return c.NoContent(http.StatusCreated)
	})

	workerWithIDGroup.PUT("", func(c echo.Context) error {
		id := c.Get("id").(uuid.UUID)
		if err := manager.workerRegistry.Renew(id); err != nil {
			return echo.NewHTTPError(
				http.StatusInternalServerError,
				fmt.Errorf("failed to renew worker's expiration time %w", err),
			)
		}

		return c.NoContent(http.StatusOK)
	})

	workerWithIDGroup.DELETE("", func(c echo.Context) error {
		var addr node.Addr
		if err := c.Bind(&addr); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid worker node address")
		}

		id := c.Get("id").(uuid.UUID)
		manager.workerRegistry.Remove(id)
		return c.NoContent(http.StatusOK)
	})

	workerGroup.POST("/message", func(c echo.Context) error {
		var message worker.Message
		if err := c.Bind(&message); err != nil {
			return echo.NewHTTPError(
				http.StatusBadRequest,
				fmt.Errorf("invalid message format: %w", err),
			)
		}

		manager.messagesQueue.Enqueue(message)
		return c.NoContent(http.StatusCreated)
	})

	// TODO(SergeyCherepiuk): Tasks with state "Scheduled" are not on the list of tasks
	e.POST("/task/run", func(c echo.Context) error {
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

	e.POST("/task/stop/:id", func(c echo.Context) error {
		id := c.Get("id").(uuid.UUID)
		manager.Stop(id)
		return c.NoContent(http.StatusCreated)
	}, parseID)

	e.GET("/task/list", func(c echo.Context) error {
		return c.JSON(http.StatusOK, manager.Tasks())
	})

	return e.Start(addr)
}

func parseID(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		id, err := uuid.Parse(c.Param("id"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid id format")
		}

		c.Set("id", id)
		return next(c)
	}
}

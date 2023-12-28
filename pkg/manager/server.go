package manager

import (
	"fmt"
	"net/http"

	"github.com/SergeyCherepiuk/fleet/pkg/node"
	"github.com/SergeyCherepiuk/fleet/pkg/task"
	"github.com/SergeyCherepiuk/fleet/pkg/worker"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func StartServer(addr string, manager *Manager) error {
	e := echo.New()
	e.HideBanner = true

	workerGroup := e.Group("/worker")
	workerWithIdGroup := workerGroup.Group("/:id", parseId)

	workerWithIdGroup.POST("", func(c echo.Context) error {
		var addr node.Addr
		if err := c.Bind(&addr); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid worker node address")
		}

		id := c.Get("id").(uuid.UUID)
		manager.AddWorker(id, addr)
		return c.NoContent(http.StatusCreated)
	})

	workerWithIdGroup.DELETE("", func(c echo.Context) error {
		var addr node.Addr
		if err := c.Bind(&addr); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid worker node address")
		}

		id := c.Get("id").(uuid.UUID)
		if err := manager.RemoveWorker(id); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err)
		}

		return c.NoContent(http.StatusOK)
	})

	workerGroup.POST("/event", func(c echo.Context) error {
		var event task.Event
		if err := c.Bind(&event); err != nil {
			return echo.NewHTTPError(
				http.StatusBadRequest,
				fmt.Errorf("invalid message format: %w", err),
			)
		}

		manager.EventsQueue.Enqueue(event)
		return c.NoContent(http.StatusCreated)
	})

	workerGroup.POST("/message", func(c echo.Context) error {
		var message worker.Message
		if err := c.Bind(&message); err != nil {
			return echo.NewHTTPError(
				http.StatusBadRequest,
				fmt.Errorf("invalid message format: %w", err),
			)
		}

		manager.WorkerMessagesQueue.Enqueue(message)
		return c.NoContent(http.StatusCreated)
	})

	e.POST("/task/run", func(c echo.Context) error {
		var t task.Task
		if err := c.Bind(&t); err != nil {
			return echo.NewHTTPError(
				http.StatusBadRequest,
				fmt.Errorf("invalid task format: %w", err),
			)
		}

		event := task.Event{Task: t, Desired: task.Running}
		manager.EventsQueue.Enqueue(event)
		return c.NoContent(http.StatusCreated)
	})

	e.POST("/task/stop/:id", func(c echo.Context) error {
		id := c.Get("id").(uuid.UUID)
		t, err := manager.Store.GetTask(id)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err)
		}

		event := task.Event{Task: t, Desired: task.Finished}
		manager.EventsQueue.Enqueue(event)
		return c.NoContent(http.StatusCreated)
	}, parseId)

	e.GET("/task/list", func(c echo.Context) error {
		return c.JSON(http.StatusOK, manager.Tasks())
	})

	return e.Start(addr)
}

func parseId(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		id, err := uuid.Parse(c.Param("id"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid id format")
		}

		c.Set("id", id)
		return next(c)
	}
}

package manager

import (
	"fmt"
	"net/http"

	httpinternal "github.com/SergeyCherepiuk/fleet/internal/http"
	"github.com/SergeyCherepiuk/fleet/pkg/httpclient"
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

	workerGroup.GET("/list", func(c echo.Context) error {
		workers := manager.Store.AllWorkers()
		infos := make([]worker.Info, 0, len(workers))
		for _, w := range workers {
			resp, err := httpclient.Get(w.Addr.String(), "/info")
			if err != nil {
				continue
			}

			var info worker.Info
			if err := httpinternal.Body(resp, &info); err != nil {
				continue
			}

			infos = append(infos, info)
		}
		return c.JSON(http.StatusOK, infos)
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
		events := manager.EventsQueue.GetAll()
		pendingTasks := make([]task.Task, 0, len(events))
		for _, event := range events {
			pendingTasks = append(pendingTasks, event.Task)
		}

		tasks := append(manager.Tasks(), pendingTasks...)
		return c.JSON(http.StatusOK, tasks)
	})

	e.GET("/task/list/:id", func(c echo.Context) error {
		id := c.Get("id").(uuid.UUID)
		return c.JSON(http.StatusOK, manager.WorkerTasks(id))
	}, parseId)

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

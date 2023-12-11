package manager

import (
	"fmt"
	"net/http"
	"time"

	"github.com/SergeyCherepiuk/fleet/pkg/node"
	"github.com/labstack/echo/v4"
)

func StartServer(addr string, manager Manager) error {
	e := echo.New()
	e.HideBanner = true

	go func() {
		for {
			fmt.Println(manager.WorkerNodes)
			time.Sleep(time.Second)
		}
	}()

	e.POST("/worker", func(c echo.Context) error {
		var addr node.Addr
		if err := c.Bind(&addr); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid worker node address")
		}

		manager.WorkerNodes = append(manager.WorkerNodes, addr)
		return c.NoContent(http.StatusCreated)
	})

	return e.Start(addr)
}

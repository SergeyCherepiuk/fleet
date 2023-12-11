package worker

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func StartServer(addr string, worker Worker) error {
	e := echo.New()
	e.HideBanner = true

	e.GET("/ping", func(c echo.Context) error {
		return c.String(http.StatusOK, "Pong from worker!")
	})

	return e.Start(addr)
}

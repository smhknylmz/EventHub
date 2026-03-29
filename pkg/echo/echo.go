package echo

import (
	"context"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	log "github.com/sirupsen/logrus"
)

func New() *echo.Echo {
	e := echo.New()
	e.HideBanner = true

	e.GET("/health", func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})

	return e
}

func Start(e *echo.Echo, addr string) {
	go func() {
		log.WithField("addr", addr).Info("starting server")
		if err := e.Start(addr); err != nil && err != http.ErrServerClosed {
			log.WithError(err).Fatal("failed to start server")
		}
	}()
}

func Shutdown(e *echo.Echo) {
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := e.Shutdown(shutdownCtx); err != nil {
		log.Fatal(err)
	}

	log.Info("server stopped")
}

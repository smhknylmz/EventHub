package echo

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
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

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := e.Shutdown(shutdownCtx); err != nil {
		log.Fatal(err)
	}

	log.Info("server stopped")
}

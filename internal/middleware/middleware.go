package middleware

import (
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	log "github.com/sirupsen/logrus"
)

const CorrelationIDHeader = "X-Correlation-ID"

func CorrelationID() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			correlationID := c.Request().Header.Get(CorrelationIDHeader)
			if correlationID == "" {
				correlationID = uuid.New().String()
			}
			c.Set("correlationId", correlationID)
			c.Response().Header().Set(CorrelationIDHeader, correlationID)
			return next(c)
		}
	}
}

func RequestLogger(logger *log.Entry) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()

			err := next(c)

			correlationID, _ := c.Get("correlationId").(string)

			logger.WithFields(log.Fields{
				"correlationId": correlationID,
				"method":        c.Request().Method,
				"path":          c.Request().URL.Path,
				"status":        c.Response().Status,
				"durationMs":    time.Since(start).Milliseconds(),
				"userAgent":     c.Request().UserAgent(),
			}).Info("request completed")

			return err
		}
	}
}

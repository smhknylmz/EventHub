package middleware

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
)

const (
	IdempotencyKeyHeader = "Idempotency-Key"
	idempotencyPrefix    = "idempotency:"
)

type cachedEntry struct {
	StatusCode int             `json:"statusCode"`
	Body       json.RawMessage `json:"body"`
}

func Idempotency(client *redis.Client, ttl time.Duration) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if c.Request().Method != http.MethodPost {
				return next(c)
			}

			key := c.Request().Header.Get(IdempotencyKeyHeader)
			if key == "" {
				return next(c)
			}

			ctx := c.Request().Context()
			redisKey := idempotencyPrefix + key

			cached, err := client.Get(ctx, redisKey).Bytes()
			if err == nil {
				var entry cachedEntry
				if json.Unmarshal(cached, &entry) == nil {
					return c.JSONBlob(entry.StatusCode, entry.Body)
				}
			}

			rec := &responseRecorder{ResponseWriter: c.Response().Writer, body: &bytes.Buffer{}, statusCode: http.StatusOK}
			c.Response().Writer = rec

			if err := next(c); err != nil {
				return err
			}

			if rec.statusCode >= 200 && rec.statusCode < 300 {
				entry, _ := json.Marshal(cachedEntry{StatusCode: rec.statusCode, Body: rec.body.Bytes()})
				client.Set(ctx, redisKey, entry, ttl)
			}

			return nil
		}
	}
}

type responseRecorder struct {
	http.ResponseWriter
	body       *bytes.Buffer
	statusCode int
}

func (r *responseRecorder) Write(b []byte) (int, error) {
	r.body.Write(b)
	return r.ResponseWriter.Write(b)
}

func (r *responseRecorder) WriteHeader(code int) {
	r.statusCode = code
	r.ResponseWriter.WriteHeader(code)
}

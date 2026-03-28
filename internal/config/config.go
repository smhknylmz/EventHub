package config

import (
	"time"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	DatabaseURL       string        `env:"DATABASE_URL,required"`
	RedisURL          string        `env:"REDIS_URL,required"`
	WebhookBaseURL    string        `env:"WEBHOOK_BASE_URL,required"`
	ServerPort        int           `env:"SERVER_PORT" envDefault:"8080"`
	LogLevel          string        `env:"LOG_LEVEL" envDefault:"info"`
	MaxRetries        int           `env:"MAX_RETRIES" envDefault:"5"`
	BackoffBase       time.Duration `env:"BACKOFF_BASE" envDefault:"1s"`
	RetryPollInterval time.Duration `env:"RETRY_POLL_INTERVAL" envDefault:"5s"`
}

func Load() (*Config, error) {
	cfg, err := env.ParseAs[Config]()
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}

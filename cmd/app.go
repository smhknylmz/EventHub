package cmd

import (
	"context"
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/smhknylmz/EventHub/internal/config"
	"github.com/smhknylmz/EventHub/internal/notification"
	pgrepo "github.com/smhknylmz/EventHub/internal/postgres"
	redisadapter "github.com/smhknylmz/EventHub/internal/redis"
	"github.com/smhknylmz/EventHub/internal/webhook"
	"github.com/smhknylmz/EventHub/internal/worker"
	"github.com/smhknylmz/EventHub/migrations"
	pkgecho "github.com/smhknylmz/EventHub/pkg/echo"
	"github.com/smhknylmz/EventHub/pkg/postgres"
	pkgredis "github.com/smhknylmz/EventHub/pkg/redis"
	pkgvalidator "github.com/smhknylmz/EventHub/pkg/validator"
)

func Execute() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(fmt.Errorf("failed to load config: %w", err))
	}

	log.SetFormatter(&log.JSONFormatter{})
	level, err := log.ParseLevel(cfg.LogLevel)
	if err != nil {
		level = log.InfoLevel
	}
	log.SetLevel(level)

	if err := postgres.Migrate(cfg.DatabaseURL, migrations.FS); err != nil {
		log.Fatal(fmt.Errorf("failed to run migrations: %w", err))
	}

	ctx := context.Background()

	pool, err := postgres.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatal(fmt.Errorf("failed to connect to database: %w", err))
	}
	defer pool.Close()

	redisClient, err := pkgredis.New(ctx, cfg.RedisURL)
	if err != nil {
		log.Fatal(fmt.Errorf("failed to connect to redis: %w", err))
	}
	defer redisClient.Close()

	queue := redisadapter.NewQueue(redisClient)
	rateLimiter := redisadapter.NewRateLimiter(redisClient)
	webhookProvider := webhook.NewProvider(cfg.WebhookBaseURL)

	notificationRepo := pgrepo.NewRepo(pool)
	notificationService := notification.NewService(notificationRepo, queue)

	processor := worker.NewProcessor(notificationRepo, rateLimiter, webhookProvider)
	dispatcher := worker.NewDispatcher(queue, processor, "worker-1")

	workerCtx, workerCancel := context.WithCancel(ctx)
	defer workerCancel()

	go dispatcher.Start(workerCtx)

	e := pkgecho.New()
	e.Validator = pkgvalidator.New()

	notificationHandler := notification.NewHandler(notificationService)
	notificationHandler.Register(e)

	pkgecho.Start(e, fmt.Sprintf(":%d", cfg.ServerPort))
}

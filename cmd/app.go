package cmd

import (
	"context"
	"fmt"

	log "github.com/sirupsen/logrus"
	pgrepo "github.com/smhknylmz/EventHub/internal/postgres"
	"github.com/smhknylmz/EventHub/internal/config"
	"github.com/smhknylmz/EventHub/internal/notification"
	"github.com/smhknylmz/EventHub/migrations"
	pkgecho "github.com/smhknylmz/EventHub/pkg/echo"
	"github.com/smhknylmz/EventHub/pkg/postgres"
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

	e := pkgecho.New()
	e.Validator = pkgvalidator.New()

	notificationRepo := pgrepo.NewRepo(pool)
	notificationService := notification.NewService(notificationRepo)
	notificationHandler := notification.NewHandler(notificationService)
	notificationHandler.Register(e)

	pkgecho.Start(e, fmt.Sprintf(":%d", cfg.ServerPort))
}

package cmd

import (
	log "github.com/sirupsen/logrus"
	"github.com/smhknylmz/EventHub/internal/notification"
	pkgecho "github.com/smhknylmz/EventHub/pkg/echo"
	pkgvalidator "github.com/smhknylmz/EventHub/pkg/validator"
)

func Execute() {
	log.SetFormatter(&log.JSONFormatter{})
	log.SetLevel(log.InfoLevel)

	e := pkgecho.New()
	e.Validator = pkgvalidator.New()

	notificationService := notification.NewService()
	notificationHandler := notification.NewHandler(notificationService)
	notificationHandler.Register(e)

	pkgecho.Start(e, ":8080")
}

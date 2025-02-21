package config

import (
	"context"
	"os"
	"os/signal"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/sirupsen/logrus"
)

func NewFiber() (*fiber.App, *logrus.Logger) {
	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})
	log := logrus.New()

	app.Use(logger.New(logger.Config{
		Output: log.Out,
	}))

	return app, log
}

func GracefulShutdown(app *fiber.App, log *logrus.Logger, shutdownTimeout time.Duration) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)

	<-quit
	log.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := app.ShutdownWithContext(ctx); err != nil {
		log.Fatalf("Failed to shutdown server: %v", err)
	}

	log.Info("Server shutdown complete")
}

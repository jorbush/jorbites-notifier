package main

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/jorbush/jorbites-notifier/config"
	"github.com/jorbush/jorbites-notifier/internal/api"
	"github.com/jorbush/jorbites-notifier/internal/logger"
	"github.com/jorbush/jorbites-notifier/internal/middleware"
	"github.com/jorbush/jorbites-notifier/internal/queue"
)

func main() {
	cleanup := logger.Init()
	defer cleanup()

	cfg := config.GetConfig()
	slog.Info("Starting jorbites-notifier service")

	mux := http.NewServeMux()
	notificationQueue := queue.NewQueue(cfg)
	notificationQueue.StartProcessing()
	notificationHandler := api.NewNotificationHandler(notificationQueue)

	mux.HandleFunc("/health", api.HealthCheckHandler)
	mux.HandleFunc("/notifications", middleware.RequireAPIKey(notificationHandler.EnqueueNotification))
	mux.HandleFunc("/queue", middleware.RequireAPIKey(notificationHandler.GetQueueStatus))

	slog.Info("Starting server", "port", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, mux); err != nil {
		slog.Error("Error starting server", "error", err)
		os.Exit(1)
	}
}

package main

import (
	"log"
	"net/http"
	"os"

	"github.com/jorbush/jorbites-notifier/config"
	"github.com/jorbush/jorbites-notifier/internal/api"
	"github.com/jorbush/jorbites-notifier/internal/queue"
)

func main() {
	cfg := config.GetConfig()
	log.SetOutput(os.Stdout)
	log.Println("Starting jorbites-notifier service")

	mux := http.NewServeMux()
	notificationQueue := queue.NewQueue()
	notificationQueue.StartProcessing()
	notificationHandler := api.NewNotificationHandler(notificationQueue)

	mux.HandleFunc("/health", api.HealthCheckHandler)
	mux.HandleFunc("/notifications", notificationHandler.EnqueueNotification)
	mux.HandleFunc("/queue", notificationHandler.GetQueueStatus)

	log.Printf("Starting server on port %s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, mux); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}

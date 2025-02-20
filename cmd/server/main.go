package main

import (
	"log"
	"net/http"

	"github.com/jorbush/jorbites-notifier/config"
	"github.com/jorbush/jorbites-notifier/internal/api"
)

func main() {
	cfg := config.GetConfig()

	mux := http.NewServeMux()

	mux.HandleFunc("/health", api.HealthCheckHandler)

	log.Printf("Starting server on port %s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, mux); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}

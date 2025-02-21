package api

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/jorbush/jorbites-notifier/internal/models"
)

func HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	response := models.APIResponse{
		Success: true,
		Data: map[string]string{
			"status": "healthy",
		},
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding response: %v", err)
	}
}

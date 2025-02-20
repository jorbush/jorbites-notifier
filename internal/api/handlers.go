package api

import (
	"encoding/json"
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
	json.NewEncoder(w).Encode(response)
}

func sendJSONResponse(w http.ResponseWriter, status int, response models.APIResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(response)
}

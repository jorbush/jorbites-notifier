package api

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/jorbush/jorbites-notifier/internal/models"
	"github.com/jorbush/jorbites-notifier/internal/queue"
)

type NotificationHandler struct {
	Queue *queue.Queue
}

func NewNotificationHandler(q *queue.Queue) *NotificationHandler {
	return &NotificationHandler{
		Queue: q,
	}
}

func (h *NotificationHandler) EnqueueNotification(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var notification models.Notification
	err := json.NewDecoder(r.Body).Decode(&notification)
	if err != nil {
		http.Error(w, "Invalid notification data: "+err.Error(), http.StatusBadRequest)
		return
	}

	if notification.Type == "" {
		http.Error(w, "Type is required", http.StatusBadRequest)
		return
	}

	h.Queue.Enqueue(notification)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(notification)
	if err != nil {
		log.Printf("Error encoding notification: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func (h *NotificationHandler) GetQueueStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	count, notifications := h.Queue.GetQueueStatus()

	response := map[string]any{
		"count":         count,
		"notifications": notifications,
	}

	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(response)
	if err != nil {
		log.Printf("Error encoding response: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

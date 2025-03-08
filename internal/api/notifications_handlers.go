package api

import (
	"encoding/json"
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
	json.NewEncoder(w).Encode(notification)
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
	json.NewEncoder(w).Encode(response)
}

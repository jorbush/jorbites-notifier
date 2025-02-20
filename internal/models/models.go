package models

import (
	"time"
)

type EmailJob struct {
	ID          string    `json:"id"`
	To          []string  `json:"to"`
	Subject     string    `json:"subject"`
	Body        string    `json:"body"`
	CreatedAt   time.Time `json:"createdAt"`
	Status      string    `json:"status"` // pending, processing, completed, failed
	RetryCount  int       `json:"retryCount"`
	LastError   string    `json:"lastError,omitempty"`
	CompletedAt time.Time `json:"completedAt,omitempty"`
}

type QueueStats struct {
	Pending    int `json:"pending"`
	Processing int `json:"processing"`
	Completed  int `json:"completed"`
	Failed     int `json:"failed"`
}

type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

package models

type NotificationStatus string

const (
	StatusPending    NotificationStatus = "pending"
	StatusProcessing NotificationStatus = "processing"
)

type NotificationType string

const (
	TypeNewComment             NotificationType = "NEW_COMMENT"
	TypeNewLike                NotificationType = "NEW_LIKE"
	TypeNewRecipe              NotificationType = "NEW_RECIPE"
	TypeNotificationsActivated NotificationType = "NOTIFICATIONS_ACTIVATED"
	TypeForgotPassword         NotificationType = "FORGOT_PASSWORD"
)

type Notification struct {
	ID        string             `json:"id,omitempty"`
	Type      NotificationType   `json:"type"`
	Status    NotificationStatus `json:"status"`
	Recipient string             `json:"recipient,omitempty"`
	Metadata  map[string]string  `json:"metadata,omitempty"`
}

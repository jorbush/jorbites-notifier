package queue

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jorbush/jorbites-notifier/config"
	"github.com/jorbush/jorbites-notifier/internal/database"
	"github.com/jorbush/jorbites-notifier/internal/email"
	"github.com/jorbush/jorbites-notifier/internal/models"
)

type Queue struct {
	notifications []models.Notification
	mutex         sync.Mutex
	processing    bool
	notifyChan    chan struct{}
	emailSender   *email.EmailSender
	mongoDB       *database.MongoDB
}

func NewQueue(cfg *config.Config) *Queue {
	mongoDB, err := database.NewMongoDB(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}

	return &Queue{
		notifications: []models.Notification{},
		notifyChan:    make(chan struct{}, 1),
		processing:    false,
		emailSender:   email.NewEmailSender(cfg),
		mongoDB:       mongoDB,
	}
}

func (q *Queue) Enqueue(notification models.Notification) {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	notification.ID = uuid.New().String()
	notification.Status = models.StatusPending

	q.notifications = append(q.notifications, notification)

	select {
	case q.notifyChan <- struct{}{}:
	default:
	}

	log.Printf("Notification %s added to queue. Queue size: %d", notification.ID, len(q.notifications))
}

func (q *Queue) GetQueueStatus() (int, []models.Notification) {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	copy := make([]models.Notification, len(q.notifications))
	for i, n := range q.notifications {
		copy[i] = n
	}

	return len(q.notifications), copy
}

func (q *Queue) StartProcessing() {
	q.mutex.Lock()
	if q.processing {
		q.mutex.Unlock()
		return
	}
	q.processing = true
	q.mutex.Unlock()

	go func() {
		for {
			select {
			case <-q.notifyChan:
			case <-time.After(5 * time.Second):
			}

			q.processNextNotification()
		}
	}()

	log.Println("Notification queue processing started")
}

func (q *Queue) processNextNotification() {
	q.mutex.Lock()

	if len(q.notifications) == 0 {
		q.mutex.Unlock()
		return
	}

	notification := q.notifications[0]
	notification.Status = models.StatusProcessing
	q.notifications[0] = notification
	q.mutex.Unlock()

	log.Printf("Processing notification %s of type %s", notification.ID, notification.Type)

	success := q.processNotificationByType(notification)

	log.Printf("Notification %s processed with success: %t", notification.ID, success)

	q.mutex.Lock()
	defer q.mutex.Unlock()

	if len(q.notifications) > 0 && q.notifications[0].ID == notification.ID {
		q.notifications = q.notifications[1:]
		log.Printf("Notification %s processed. Queue size: %d", notification.ID, len(q.notifications))
	}
}

func (q *Queue) processNotificationByType(notification models.Notification) bool {
	switch notification.Type {
	case models.TypeNewRecipe:
		return q.processNewRecipeNotification(notification)
	case models.TypeNewBlog:
		return q.processNewBlogNotification(notification)
	case models.TypeNewComment, models.TypeNewLike, models.TypeNotificationsActivated, models.TypeForgotPassword:
		success, err := q.emailSender.SendNotificationEmail(notification)
		if err != nil {
			log.Printf("Error sending email for notification %s: %v", notification.ID, err)
			return false
		}
		return success
	case models.TypeMentionInComment:
		return q.processMentionInCommentNotification(notification)
	default:
		log.Printf("Unknown notification type: %s", notification.Type)
		return false
	}
}

func (q *Queue) processNewRecipeNotification(notification models.Notification) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	users, err := q.mongoDB.GetUsersWithNotificationsEnabled(ctx)
	if err != nil {
		log.Printf("Error fetching users for notification %s: %v", notification.ID, err)
		return false
	}

	log.Printf("Sending new recipe notification to %d users with notifications enabled", len(users))

	successCount := 0
	failCount := 0

	for _, user := range users {
		userNotification := models.Notification{
			ID:        uuid.New().String(),
			Type:      notification.Type,
			Status:    models.StatusProcessing,
			Recipient: user.Email,
			Metadata:  notification.Metadata,
		}

		success, err := q.emailSender.SendNotificationEmail(userNotification)
		if err != nil {
			log.Printf("Error sending email to %s: %v", user.Email, err)
			failCount++
			continue
		}

		if success {
			successCount++
		} else {
			failCount++
		}

		// Add a small delay to avoid overwhelming the SMTP server
		time.Sleep(100 * time.Millisecond)
	}

	log.Printf("New recipe notification results: %d successful, %d failed", successCount, failCount)
	return successCount > 0
}

func (q *Queue) processMentionInCommentNotification(notification models.Notification) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	users, err := q.mongoDB.GetUsersMentionedInComment(ctx, notification.Metadata["mentionedUsers"], notification.Recipient)
	if err != nil {
		log.Printf("Error fetching users for mention notification %s: %v", notification.ID, err)
		return false
	}

	log.Printf("Sending mention in comment notification to %d users", len(users))

	successCount := 0
	failCount := 0

	for _, user := range users {
		userNotification := models.Notification{
			ID:        uuid.New().String(),
			Type:      notification.Type,
			Status:    models.StatusProcessing,
			Recipient: user.Email,
			Metadata:  notification.Metadata,
		}

		success, err := q.emailSender.SendNotificationEmail(userNotification)
		if err != nil {
			log.Printf("Error sending email to %s: %v", user.Email, err)
			failCount++
			continue
		}

		if success {
			successCount++
		} else {
			failCount++
		}

		time.Sleep(100 * time.Millisecond)
	}

	log.Printf("Mention in comment notification results: %d successful, %d failed", successCount, failCount)
	return successCount > 0
}

func (q *Queue) processNewBlogNotification(notification models.Notification) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	users, err := q.mongoDB.GetUsersWithNotificationsEnabled(ctx)
	if err != nil {
		log.Printf("Error fetching users for notification %s: %v", notification.ID, err)
		return false
	}

	log.Printf("Sending new blog notification to %d users with notifications enabled", len(users))

	successCount := 0
	failCount := 0

	for _, user := range users {
		userNotification := models.Notification{
			ID:        uuid.New().String(),
			Type:      notification.Type,
			Status:    models.StatusProcessing,
			Recipient: user.Email,
			Metadata:  notification.Metadata,
		}

		success, err := q.emailSender.SendNotificationEmail(userNotification)
		if err != nil {
			log.Printf("Error sending email to %s: %v", user.Email, err)
			failCount++
			continue
		}

		if success {
			successCount++
		} else {
			failCount++
		}

		time.Sleep(100 * time.Millisecond)
	}

	log.Printf("New blog notification results: %d successful, %d failed", successCount, failCount)
	return successCount > 0
}

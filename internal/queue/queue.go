package queue

import (
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jorbush/jorbites-notifier/internal/models"
)

type Queue struct {
	notifications []models.Notification
	mutex         sync.Mutex
	processing    bool
	notifyChan    chan struct{}
}

func NewQueue() *Queue {
	return &Queue{
		notifications: []models.Notification{},
		notifyChan:    make(chan struct{}, 1),
		processing:    false,
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
	q.mutex.Unlock()

	log.Printf("Processing notification %s of type %s", notification.ID, notification.Type)

	success := processNotificationByType(notification)

	log.Printf("Notification %s processed with success: %t", notification.ID, success)

	q.mutex.Lock()
	defer q.mutex.Unlock()

	if len(q.notifications) > 0 && q.notifications[0].ID == notification.ID {
		q.notifications = q.notifications[1:]
		log.Printf("Notification %s processed. Queue size: %d", notification.ID, len(q.notifications))
	}
}

func processNotificationByType(notification models.Notification) bool {
	time.Sleep(500 * time.Millisecond)

	switch notification.Type {
	case models.TypeNewComment:
		return sendEmailNotification(notification)
	case models.TypeNewLike:
		return sendEmailNotification(notification)
	case models.TypeNewRecipe:
		return sendEmailNotification(notification)
	case models.TypeNotificationsActived:
		return sendEmailNotification(notification)
	default:
		log.Printf("Unknown notification type: %s", notification.Type)
		return false
	}
}

func sendEmailNotification(notification models.Notification) bool {
	// TODO: Implement email sending
	log.Printf("Sending email notification to %s for type %s",
		notification.Recipient, notification.Type)
	// simulate processing time
	time.Sleep(3000 * time.Millisecond)
	return true
}

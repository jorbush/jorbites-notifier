package queue

import (
	"context"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jorbush/jorbites-notifier/config"
	"github.com/jorbush/jorbites-notifier/internal/database"
	"github.com/jorbush/jorbites-notifier/internal/email"
	"github.com/jorbush/jorbites-notifier/internal/models"
	"github.com/jorbush/jorbites-notifier/internal/push"
)

type Queue struct {
	notifications []models.Notification
	mutex         sync.Mutex
	processing    bool
	notifyChan    chan struct{}
	emailSender   *email.EmailSender
	pushSender    *push.PushSender
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
		pushSender:    push.NewPushSender(cfg, mongoDB),
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
	case models.TypeForgotPassword:
		// 1. Send Email (Transactional - ignore preferences)
		success, err := q.emailSender.SendNotificationEmail(notification)
		if err != nil {
			log.Printf("Error sending email for notification %s: %v", notification.ID, err)
			return false
		}
		return success

	case models.TypeNewComment, models.TypeNewLike, models.TypeNotificationsActivated:
		// 1. Lookup User first (needed for both Push and Email preference)
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		user, err := q.mongoDB.GetUserByEmail(ctx, notification.Recipient)
		if err != nil {
			log.Printf("Error fetching user for recipient %s: %v", notification.Recipient, err)
			return false
		}

		// 2. Send Email if enabled
		var success bool = true
		if user.EmailNotifications {
			var err error
			success, err = q.emailSender.SendNotificationEmail(notification)
			if err != nil {
				log.Printf("Error sending email for notification %s: %v", notification.ID, err)
				// Log error but continue to push? Or return false?
				// Usually we want to try push even if email fails, but return valid status.
				// Let's count success based on email if attempted.
				success = false
			}
		} else {
			log.Printf("Skipping email for %s (notifications disabled)", notification.Recipient)
		}

		// 3. Send Push Notification
		userID := user.ID.Hex()

		var title, message, url string
		switch notification.Type {
		case models.TypeNewLike:
			title = "New Like"
			// Metadata: likedBy, recipeId
			likedBy := notification.Metadata["likedBy"]
			recipeId := notification.Metadata["recipeId"]

			message = "Someone liked your recipe"
			if likedBy != "" {
				message = likedBy + " liked your recipe"
			}
			url = "/recipes/" + recipeId
		case models.TypeNewComment:
			title = "New Comment"
			// Metadata: commentId, authorName, recipeId
			authorName := notification.Metadata["authorName"]
			recipeId := notification.Metadata["recipeId"]

			message = "New comment on your recipe"
			if authorName != "" {
				message = authorName + " commented on your recipe"
			}
			url = "/recipes/" + recipeId
		case models.TypeNotificationsActivated:
			title = "Notifications Activated"
			message = "You have successfully activated notifications"
			url = "/settings/notifications"
		}

		if title != "" {
			log.Printf("Sending push notification '%s' to user %s", title, userID)
			q.sendPushToUsers([]string{userID}, notification, title, message, url)
		} else {
			log.Printf("No push notification title set for type %s", notification.Type)
		}

		return success
	case models.TypeMentionInComment:
		return q.processMentionInCommentNotification(notification)
	default:
		log.Printf("Unknown notification type: %s", notification.Type)
		return false
	}
}

// Helper to broadcast push notifications to all subscribers
func (q *Queue) broadcastPushNotification(notification models.Notification, title, message, url string) {

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	subs, err := q.mongoDB.GetAllPushSubscriptions(ctx)
	if err != nil {
		log.Printf("Error fetching push subscriptions for broadcast: %v", err)
		return
	}

	for _, sub := range subs {
		go func(s models.PushSubscription) {
			if err := q.pushSender.SendNotification(s, title, message, url); err != nil {
				log.Printf("Error sending push to %s: %v", s.ID.Hex(), err)
			}
		}(sub)
	}
}

// Helper to send push to specific users
func (q *Queue) sendPushToUsers(userIDs []string, notification models.Notification, title, message, url string) {
	if len(userIDs) == 0 {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	subs, err := q.mongoDB.GetPushSubscriptionsForUsers(ctx, userIDs)
	if err != nil {
		log.Printf("Error fetching push subscriptions for users: %v", err)
		return
	}

	log.Printf("Found %d push subscriptions for users %v", len(subs), userIDs)

	for _, sub := range subs {
		go func(s models.PushSubscription) {
			if err := q.pushSender.SendNotification(s, title, message, url); err != nil {
				log.Printf("Error sending push to %s: %v", s.ID.Hex(), err)
			} else {
				log.Printf("Push sent to subscription %s", s.ID.Hex())
			}
		}(sub)
	}
}

func (q *Queue) processNewRecipeNotification(notification models.Notification) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 1. Send Emails
	users, err := q.mongoDB.GetUsersWithNotificationsEnabled(ctx)
	var emailSuccess bool
	if err != nil {
		log.Printf("Error fetching users for notification %s: %v", notification.ID, err)
		emailSuccess = false
	} else {
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
			time.Sleep(100 * time.Millisecond)
		}
		log.Printf("New recipe email results: %d successful, %d failed", successCount, failCount)
		emailSuccess = successCount > 0
	}

	// 2. Send Push Notifications
	recipeName := notification.Metadata["recipeName"]
	q.broadcastPushNotification(notification, "New Recipe!", "New recipe available: "+recipeName, "/recipes/"+notification.Metadata["slug"])

	return emailSuccess
}

func (q *Queue) processMentionInCommentNotification(notification models.Notification) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 1. Send Emails
	users, err := q.mongoDB.GetUsersMentionedInComment(ctx, notification.Metadata["mentionedUsers"], notification.Recipient)
	var emailSuccess bool
	if err != nil {
		log.Printf("Error fetching users for mention notification %s: %v", notification.ID, err)
		emailSuccess = false
	} else {
		// Existing email logic
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
		log.Printf("Mention in comment email results: %d successful, %d failed", successCount, failCount)
		emailSuccess = successCount > 0
	}

	// 2. Send Push Notifications
	mentionedUserIDsStr := notification.Metadata["mentionedUsers"]
	if mentionedUserIDsStr != "" {
		ids := strings.Split(mentionedUserIDsStr, ",")
		q.sendPushToUsers(ids, notification, "You were mentioned!", "You were mentioned in a comment.", "/recipes/"+notification.Metadata["recipeId"])
	}

	return emailSuccess
}

func (q *Queue) processNewBlogNotification(notification models.Notification) bool {
	// Similar logic to NewRecipe
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	users, err := q.mongoDB.GetUsersWithNotificationsEnabled(ctx)
	var emailSuccess bool
	if err != nil {
		log.Printf("Error fetching users for notification %s: %v", notification.ID, err)
		emailSuccess = false
	} else {
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
		emailSuccess = successCount > 0
	}

	// 2. Push
	postTitle := notification.Metadata["title"]
	q.broadcastPushNotification(notification, "New Blog Post!", postTitle, "/blog/"+notification.Metadata["blog_id"])

	return emailSuccess
}

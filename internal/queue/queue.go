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
	"github.com/jorbush/jorbites-notifier/internal/i18n"
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
	case models.TypeNewEvent:
		return q.processNewEventNotification(notification)
	case models.TypeEventEndingSoon:
		return q.processEventEndingSoonNotification(notification)
	case models.TypeForgotPassword:
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		user, err := q.mongoDB.GetUserByEmail(ctx, notification.Recipient)
		var language string = "es"
		if err != nil {
			log.Printf("Error fetching user for recipient %s: %v (using default language)", notification.Recipient, err)
		} else {
			language = i18n.GetUserLanguage(user)
		}

		success, err := q.emailSender.SendNotificationEmail(notification, language)
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

		language := i18n.GetUserLanguage(user)

		var success bool = true
		if user.EmailNotifications {
			var err error
			success, err = q.emailSender.SendNotificationEmail(notification, language)
			if err != nil {
				log.Printf("Error sending email for notification %s: %v", notification.ID, err)
				success = false
			}
		} else {
			log.Printf("Skipping email for %s (notifications disabled)", notification.Recipient)
		}

		userID := user.ID.Hex()

		var title, message, url string
		switch notification.Type {
		case models.TypeNewLike:
			recipeId := notification.Metadata["recipeId"]
			pushTexts := i18n.GetPushNotificationText(models.TypeNewLike, language, notification.Metadata)
			title = pushTexts.Title
			message = pushTexts.Message
			url = "/recipes/" + recipeId
		case models.TypeNewComment:
			recipeId := notification.Metadata["recipeId"]
			pushTexts := i18n.GetPushNotificationText(models.TypeNewComment, language, notification.Metadata)
			title = pushTexts.Title
			message = pushTexts.Message
			url = "/recipes/" + recipeId
		case models.TypeNotificationsActivated:
			pushTexts := i18n.GetPushNotificationText(models.TypeNotificationsActivated, language, notification.Metadata)
			title = pushTexts.Title
			message = pushTexts.Message
			url = "/settings/notifications"
		}

		if title != "" {
			log.Printf("Sending push notification '%s' to user %s", title, userID)

			subsCtx, subsCancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer subsCancel()

			subs, err := q.mongoDB.GetPushSubscriptionsForUsers(subsCtx, []string{userID})
			if err != nil {
				log.Printf("Error fetching push subscriptions for user %s: %v", userID, err)
			} else {
				log.Printf("Found %d push subscriptions for user %s", len(subs), userID)
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

func (q *Queue) broadcastPushNotificationMultiLang(notification models.Notification, url string) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	subs, err := q.mongoDB.GetAllPushSubscriptions(ctx)
	if err != nil {
		log.Printf("Error fetching push subscriptions for broadcast: %v", err)
		return
	}

	for _, sub := range subs {
		go func(s models.PushSubscription) {
			userCtx, userCancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer userCancel()

			user, err := q.mongoDB.GetUserByID(userCtx, s.UserID.Hex())
			var language string = "es"
			if err != nil {
				log.Printf("Error fetching user %s for push notification: %v (using default language)", s.UserID.Hex(), err)
			} else {
				language = i18n.GetUserLanguage(user)
			}

			pushTexts := i18n.GetPushNotificationText(notification.Type, language, notification.Metadata)

			if err := q.pushSender.SendNotification(s, pushTexts.Title, pushTexts.Message, url); err != nil {
				log.Printf("Error sending push to %s: %v", s.ID.Hex(), err)
			}
		}(sub)
	}
}

func (q *Queue) sendPushToUsersMultiLang(userIDs []string, notification models.Notification, url string) {
	if len(userIDs) == 0 {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	subs, err := q.mongoDB.GetPushSubscriptionsForUsers(ctx, userIDs)
	if err != nil {
		log.Printf("Error fetching push subscriptions for users: %v", err)
		return
	}

	log.Printf("Found %d push subscriptions for users %v", len(subs), userIDs)

	for _, sub := range subs {
		go func(s models.PushSubscription) {
			userCtx, userCancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer userCancel()

			user, err := q.mongoDB.GetUserByID(userCtx, s.UserID.Hex())
			var language string = "es"
			if err != nil {
				log.Printf("Error fetching user %s for push notification: %v (using default language)", s.UserID.Hex(), err)
			} else {
				language = i18n.GetUserLanguage(user)
			}

			pushTexts := i18n.GetPushNotificationText(notification.Type, language, notification.Metadata)

			if err := q.pushSender.SendNotification(s, pushTexts.Title, pushTexts.Message, url); err != nil {
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

			language := i18n.GetUserLanguage(&user)

			success, err := q.emailSender.SendNotificationEmail(userNotification, language)
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

	q.broadcastPushNotificationMultiLang(notification, "/recipes/"+notification.Metadata["slug"])

	return emailSuccess
}

func (q *Queue) processMentionInCommentNotification(notification models.Notification) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	users, err := q.mongoDB.GetUsersMentionedInComment(ctx, notification.Metadata["mentionedUsers"], notification.Recipient)
	var emailSuccess bool
	if err != nil {
		log.Printf("Error fetching users for mention notification %s: %v", notification.ID, err)
		emailSuccess = false
	} else {
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
			language := i18n.GetUserLanguage(&user)

			success, err := q.emailSender.SendNotificationEmail(userNotification, language)
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

	mentionedUserIDsStr := notification.Metadata["mentionedUsers"]
	if mentionedUserIDsStr != "" {
		ids := strings.Split(mentionedUserIDsStr, ",")
		q.sendPushToUsersMultiLang(ids, notification, "/recipes/"+notification.Metadata["recipeId"])
	}

	return emailSuccess
}

func (q *Queue) processNewBlogNotification(notification models.Notification) bool {
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
			language := i18n.GetUserLanguage(&user)
			success, err := q.emailSender.SendNotificationEmail(userNotification, language)
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

	q.broadcastPushNotificationMultiLang(notification, "/blog/"+notification.Metadata["blog_id"])

	return emailSuccess
}

func (q *Queue) processNewEventNotification(notification models.Notification) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	users, err := q.mongoDB.GetUsersWithNotificationsEnabled(ctx)
	var emailSuccess bool
	if err != nil {
		log.Printf("Error fetching users for notification %s: %v", notification.ID, err)
		emailSuccess = false
	} else {
		log.Printf("Sending new event notification to %d users with notifications enabled", len(users))
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
			language := i18n.GetUserLanguage(&user)
			success, err := q.emailSender.SendNotificationEmail(userNotification, language)
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
		log.Printf("New event notification results: %d successful, %d failed", successCount, failCount)
		emailSuccess = successCount > 0
	}

	q.broadcastPushNotificationMultiLang(notification, "/events/"+notification.Metadata["eventId"])

	return emailSuccess
}

func (q *Queue) processEventEndingSoonNotification(notification models.Notification) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	users, err := q.mongoDB.GetUsersWithNotificationsEnabled(ctx)
	var emailSuccess bool
	if err != nil {
		log.Printf("Error fetching users for notification %s: %v", notification.ID, err)
		emailSuccess = false
	} else {
		log.Printf("Sending event ending soon notification to %d users with notifications enabled", len(users))
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
			language := i18n.GetUserLanguage(&user)
			success, err := q.emailSender.SendNotificationEmail(userNotification, language)
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
		log.Printf("Event ending soon notification results: %d successful, %d failed", successCount, failCount)
		emailSuccess = successCount > 0
	}

	q.broadcastPushNotificationMultiLang(notification, "/events/"+notification.Metadata["eventId"])

	return emailSuccess
}

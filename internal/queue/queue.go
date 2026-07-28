package queue

import (
	"context"
	"log/slog"
	"os"
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
		slog.Error("Failed to connect to MongoDB", "error", err)
		os.Exit(1)
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

	slog.Info("Notification added to queue", "notificationId", notification.ID, "queueSize", len(q.notifications))
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

	slog.Info("Notification queue processing started")
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

	slog.Info("Processing notification", "notificationId", notification.ID, "type", notification.Type)

	success := q.processNotificationByType(notification)

	slog.Info("Notification processed", "notificationId", notification.ID, "success", success)

	q.mutex.Lock()
	defer q.mutex.Unlock()

	if len(q.notifications) > 0 && q.notifications[0].ID == notification.ID {
		q.notifications = q.notifications[1:]
		slog.Info("Notification processed", "notificationId", notification.ID, "queueSize", len(q.notifications))
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
		language := "es"
		if err != nil {
			slog.Error("Error fetching user, using default language", "recipient", notification.Recipient, "error", err)
		} else {
			language = i18n.GetUserLanguage(user)
		}

		success, err := q.emailSender.SendNotificationEmail(notification, language)
		if err != nil {
			slog.Error("Error sending email for notification", "notificationId", notification.ID, "error", err)
			return false
		}
		return success

	case models.TypeNewComment, models.TypeNewLike, models.TypeNotificationsActivated, models.TypeQuestFulfilled:
		// 1. Lookup User first (needed for both Push and Email preference)
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		user, err := q.mongoDB.GetUserByEmail(ctx, notification.Recipient)
		if err != nil {
			slog.Error("Error fetching user", "recipient", notification.Recipient, "error", err)
			return false
		}

		language := i18n.GetUserLanguage(user)

		if !user.IsNotificationCategoryEnabled(notification.Type) {
			slog.Info("Skipping notification (category disabled)", "recipient", notification.Recipient, "type", notification.Type)
			return true
		}

		success := true
		if user.EmailNotifications {
			var err error
			success, err = q.emailSender.SendNotificationEmail(notification, language)
			if err != nil {
				slog.Error("Error sending email for notification", "notificationId", notification.ID, "error", err)
				success = false
			}
		} else {
			slog.Info("Skipping email notification (disabled)", "recipient", notification.Recipient)
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
			url = "/"
		case models.TypeQuestFulfilled:
			pushTexts := i18n.GetPushNotificationText(models.TypeQuestFulfilled, language, notification.Metadata)
			title = pushTexts.Title
			message = pushTexts.Message
			url = "/quests/" + notification.Metadata["questId"]
		}

		if title != "" {
			slog.Info("Sending push notification", "title", title, "userId", userID)

			subsCtx, subsCancel := context.WithTimeout(context.Background(), 10*time.Second)
			subs, err := q.mongoDB.GetPushSubscriptionsForUsers(subsCtx, []string{userID})
			subsCancel()
			if err != nil {
				slog.Error("Error fetching push subscriptions for user", "userId", userID, "error", err)
			} else {
				slog.Info("Found push subscriptions for user", "count", len(subs), "userId", userID)
				for _, sub := range subs {
					go func(s models.PushSubscription) {
						if err := q.pushSender.SendNotification(s, title, message, url); err != nil {
							slog.Error("Error sending push notification", "subscriptionId", s.ID.Hex(), "error", err)
						} else {
							slog.Info("Push sent to subscription", "subscriptionId", s.ID.Hex())
						}
					}(sub)
				}
			}
		} else {
			slog.Warn("No push notification title set for type", "type", notification.Type)
		}

		return success
	case models.TypeMentionInComment:
		return q.processMentionInCommentNotification(notification)
	case models.TypeNewQuest:
		return q.processNewQuestNotification(notification)
	case models.TypeNewChallenge:
		return q.processNewChallengeNotification(notification)
	case models.TypeNewBadge:
		return q.processNewBadgeNotification(notification)
	case models.TypeVerified:
		return q.processVerifiedNotification(notification)
	case models.TypeNewVotation:
		return q.processNewVotationNotification(notification)
	case models.TypeVotationResult:
		return q.processVotationResultNotification(notification)
	default:
		slog.Warn("Unknown notification type", "type", notification.Type)
		return false
	}
}

func (q *Queue) broadcastPushNotificationMultiLang(notification models.Notification, url string) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	subs, err := q.mongoDB.GetAllPushSubscriptions(ctx)
	if err != nil {
		slog.Error("Error fetching push subscriptions for broadcast", "error", err)
		return
	}

	for _, sub := range subs {
		go func(s models.PushSubscription) {
			userCtx, userCancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer userCancel()

			user, err := q.mongoDB.GetUserByID(userCtx, s.UserID.Hex())
			language := "es"
			if err != nil {
				slog.Error("Error fetching user for push notification, using default language", "userId", s.UserID.Hex(), "error", err)
			} else {
				if !user.IsNotificationCategoryEnabled(notification.Type) {
					slog.Info("Skipping push notification (category disabled)", "userId", s.UserID.Hex(), "type", notification.Type)
					return
				}
				language = i18n.GetUserLanguage(user)
			}

			pushTexts := i18n.GetPushNotificationText(notification.Type, language, notification.Metadata)

			if err := q.pushSender.SendNotification(s, pushTexts.Title, pushTexts.Message, url); err != nil {
				slog.Error("Error sending push notification", "subscriptionId", s.ID.Hex(), "error", err)
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
		slog.Error("Error fetching push subscriptions for users", "error", err)
		return
	}

	slog.Info("Found push subscriptions for users", "count", len(subs), "userIds", userIDs)

	for _, sub := range subs {
		go func(s models.PushSubscription) {
			userCtx, userCancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer userCancel()

			user, err := q.mongoDB.GetUserByID(userCtx, s.UserID.Hex())
			language := "es"
			if err != nil {
				slog.Error("Error fetching user for push notification, using default language", "userId", s.UserID.Hex(), "error", err)
			} else {
				if !user.IsNotificationCategoryEnabled(notification.Type) {
					slog.Info("Skipping push notification (category disabled)", "userId", s.UserID.Hex(), "type", notification.Type)
					return
				}
				language = i18n.GetUserLanguage(user)
			}

			pushTexts := i18n.GetPushNotificationText(notification.Type, language, notification.Metadata)

			if err := q.pushSender.SendNotification(s, pushTexts.Title, pushTexts.Message, url); err != nil {
				slog.Error("Error sending push notification", "subscriptionId", s.ID.Hex(), "error", err)
			} else {
				slog.Info("Push sent to subscription", "subscriptionId", s.ID.Hex())
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
		slog.Error("Error fetching users for notification", "notificationId", notification.ID, "error", err)
		emailSuccess = false
	} else {
		slog.Info("Sending new recipe notification", "count", len(users))
		successCount := 0
		failCount := 0

		for _, user := range users {
			if !user.IsNotificationCategoryEnabled(notification.Type) {
				slog.Info("Skipping email notification (category disabled)", "recipient", user.Email, "type", notification.Type)
				continue
			}

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
				slog.Error("Error sending email", "email", user.Email, "error", err)
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
		slog.Info("New recipe email results", "successful", successCount, "failed", failCount)
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
		slog.Error("Error fetching users for mention notification", "notificationId", notification.ID, "error", err)
		emailSuccess = false
	} else {
		slog.Info("Sending mention in comment notification", "count", len(users))
		successCount := 0
		failCount := 0
		for _, user := range users {
			if !user.IsNotificationCategoryEnabled(notification.Type) {
				slog.Info("Skipping email notification (category disabled)", "recipient", user.Email, "type", notification.Type)
				continue
			}

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
				slog.Error("Error sending email", "email", user.Email, "error", err)
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
		slog.Info("Mention in comment email results", "successful", successCount, "failed", failCount)
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
		slog.Error("Error fetching users for notification", "notificationId", notification.ID, "error", err)
		emailSuccess = false
	} else {
		slog.Info("Sending new blog notification", "count", len(users))
		successCount := 0
		failCount := 0
		for _, user := range users {
			if !user.IsNotificationCategoryEnabled(notification.Type) {
				slog.Info("Skipping email notification (category disabled)", "recipient", user.Email, "type", notification.Type)
				continue
			}

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
				slog.Error("Error sending email", "email", user.Email, "error", err)
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
		slog.Info("New blog notification results", "successful", successCount, "failed", failCount)
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
		slog.Error("Error fetching users for notification", "notificationId", notification.ID, "error", err)
		emailSuccess = false
	} else {
		slog.Info("Sending new event notification", "count", len(users))
		successCount := 0
		failCount := 0
		for _, user := range users {
			if !user.IsNotificationCategoryEnabled(notification.Type) {
				slog.Info("Skipping email notification (category disabled)", "recipient", user.Email, "type", notification.Type)
				continue
			}

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
				slog.Error("Error sending email", "email", user.Email, "error", err)
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
		slog.Info("New event notification results", "successful", successCount, "failed", failCount)
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
		slog.Error("Error fetching users for notification", "notificationId", notification.ID, "error", err)
		emailSuccess = false
	} else {
		slog.Info("Sending event ending soon notification", "count", len(users))
		successCount := 0
		failCount := 0
		for _, user := range users {
			if !user.IsNotificationCategoryEnabled(notification.Type) {
				slog.Info("Skipping email notification (category disabled)", "recipient", user.Email, "type", notification.Type)
				continue
			}

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
				slog.Error("Error sending email", "email", user.Email, "error", err)
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
		slog.Info("Event ending soon notification results", "successful", successCount, "failed", failCount)
		emailSuccess = successCount > 0
	}

	q.broadcastPushNotificationMultiLang(notification, "/events/"+notification.Metadata["eventId"])

	return emailSuccess
}

func (q *Queue) processNewQuestNotification(notification models.Notification) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	users, err := q.mongoDB.GetUsersWithNotificationsEnabled(ctx)
	var emailSuccess bool
	if err != nil {
		slog.Error("Error fetching users for notification", "notificationId", notification.ID, "error", err)
		emailSuccess = false
	} else {
		slog.Info("Sending new quest notification", "count", len(users))
		successCount := 0
		failCount := 0
		for _, user := range users {
			if !user.IsNotificationCategoryEnabled(notification.Type) {
				slog.Info("Skipping email notification (category disabled)", "recipient", user.Email, "type", notification.Type)
				continue
			}

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
				slog.Error("Error sending email", "email", user.Email, "error", err)
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
		slog.Info("New quest notification results", "successful", successCount, "failed", failCount)
		emailSuccess = successCount > 0
	}

	q.broadcastPushNotificationMultiLang(notification, "/quests/"+notification.Metadata["questId"])

	return emailSuccess
}

func (q *Queue) processNewChallengeNotification(notification models.Notification) bool {
	if desc, ok := notification.Metadata["description"]; ok {
		notification.Metadata["description"] = strings.ReplaceAll(desc, "_", " ")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	users, err := q.mongoDB.GetUsersWithNotificationsEnabled(ctx)
	var emailSuccess bool
	if err != nil {
		slog.Error("Error fetching users for notification", "notificationId", notification.ID, "error", err)
		emailSuccess = false
	} else {
		slog.Info("Sending new challenge notification", "count", len(users))
		successCount := 0
		failCount := 0
		for _, user := range users {
			if !user.IsNotificationCategoryEnabled(notification.Type) {
				slog.Info("Skipping email notification (category disabled)", "recipient", user.Email, "type", notification.Type)
				continue
			}
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
				slog.Error("Error sending email", "email", user.Email, "error", err)
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
		slog.Info("New challenge notification results", "successful", successCount, "failed", failCount)
		emailSuccess = successCount > 0
	}

	q.broadcastPushNotificationMultiLang(notification, "/events")

	return emailSuccess
}

func (q *Queue) processNewVotationNotification(notification models.Notification) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	users, err := q.mongoDB.GetUsersWithNotificationsEnabled(ctx)
	var emailSuccess bool
	if err != nil {
		slog.Error("Error fetching users for notification", "notificationId", notification.ID, "error", err)
		emailSuccess = false
	} else {
		slog.Info("Sending new votation notification", "count", len(users))
		successCount := 0
		failCount := 0
		for _, user := range users {
			if !user.IsNotificationCategoryEnabled(notification.Type) {
				slog.Info("Skipping email notification (category disabled)", "recipient", user.Email, "type", notification.Type)
				continue
			}
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
				slog.Error("Error sending email", "email", user.Email, "error", err)
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
		slog.Info("New votation notification results", "successful", successCount, "failed", failCount)
		emailSuccess = successCount > 0
	}

	q.broadcastPushNotificationMultiLang(notification, "/events")

	return emailSuccess
}

func (q *Queue) processVotationResultNotification(notification models.Notification) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	users, err := q.mongoDB.GetUsersWithNotificationsEnabled(ctx)
	var emailSuccess bool
	if err != nil {
		slog.Error("Error fetching users for notification", "notificationId", notification.ID, "error", err)
		emailSuccess = false
	} else {
		slog.Info("Sending votation result notification", "count", len(users))
		successCount := 0
		failCount := 0
		for _, user := range users {
			if !user.IsNotificationCategoryEnabled(notification.Type) {
				slog.Info("Skipping email notification (category disabled)", "recipient", user.Email, "type", notification.Type)
				continue
			}
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
				slog.Error("Error sending email", "email", user.Email, "error", err)
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
		slog.Info("Votation result notification results", "successful", successCount, "failed", failCount)
		emailSuccess = successCount > 0
	}

	q.broadcastPushNotificationMultiLang(notification, "/events")

	return emailSuccess
}

func (q *Queue) processNewBadgeNotification(notification models.Notification) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	user, err := q.mongoDB.GetUserByEmail(ctx, notification.Recipient)
	if err != nil {
		slog.Error("Error fetching user", "recipient", notification.Recipient, "error", err)
		return false
	}

	language := i18n.GetUserLanguage(user)
	if !user.IsNotificationCategoryEnabled(notification.Type) {
		slog.Info("Skipping notification (category disabled)", "recipient", notification.Recipient, "type", notification.Type)
		return true
	}
	success := true

	if notification.Metadata == nil {
		notification.Metadata = make(map[string]string)
	}
	notification.Metadata["userId"] = user.ID.Hex()
	if badge_name, ok := notification.Metadata["badgeName"]; ok && strings.Contains(badge_name, "_") {
		notification.Metadata["badgeName"] = strings.ToUpper(strings.ReplaceAll(badge_name, "_", " "))
	}
	if user.EmailNotifications {
		success, err = q.emailSender.SendNotificationEmail(notification, language)
		if err != nil {
			slog.Error("Error sending email for notification", "notificationId", notification.ID, "error", err)
			success = false
		}
	} else {
		slog.Info("Skipping email notification (disabled)", "recipient", notification.Recipient)
	}

	pushTexts := i18n.GetPushNotificationText(notification.Type, language, notification.Metadata)
	userID := user.ID.Hex()
	url := "/profile/" + userID

	if pushTexts.Title != "" {
		slog.Info("Sending push notification", "title", pushTexts.Title, "userId", userID)

		subsCtx, subsCancel := context.WithTimeout(context.Background(), 10*time.Second)
		subs, err := q.mongoDB.GetPushSubscriptionsForUsers(subsCtx, []string{userID})
		subsCancel()
		if err != nil {
			slog.Error("Error fetching push subscriptions for user", "userId", userID, "error", err)
		} else {
			slog.Info("Found push subscriptions for user", "count", len(subs), "userId", userID)
			for _, sub := range subs {
				go func(s models.PushSubscription) {
					if err := q.pushSender.SendNotification(s, pushTexts.Title, pushTexts.Message, url); err != nil {
						slog.Error("Error sending push notification", "subscriptionId", s.ID.Hex(), "error", err)
					} else {
						slog.Info("Push sent to subscription", "subscriptionId", s.ID.Hex())
					}
				}(sub)
			}
		}
	} else {
		slog.Warn("No push notification title set for type", "type", notification.Type)
	}

	return success
}

func (q *Queue) processVerifiedNotification(notification models.Notification) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	user, err := q.mongoDB.GetUserByEmail(ctx, notification.Recipient)
	if err != nil {
		slog.Error("Error fetching user", "recipient", notification.Recipient, "error", err)
		return false
	}

	language := i18n.GetUserLanguage(user)
	if !user.IsNotificationCategoryEnabled(notification.Type) {
		slog.Info("Skipping notification (category disabled)", "recipient", notification.Recipient, "type", notification.Type)
		return true
	}
	success := true

	if notification.Metadata == nil {
		notification.Metadata = make(map[string]string)
	}
	notification.Metadata["userId"] = user.ID.Hex()

	if user.EmailNotifications {
		success, err = q.emailSender.SendNotificationEmail(notification, language)
		if err != nil {
			slog.Error("Error sending email for notification", "notificationId", notification.ID, "error", err)
			success = false
		}
	} else {
		slog.Info("Skipping email notification (disabled)", "recipient", notification.Recipient)
	}

	pushTexts := i18n.GetPushNotificationText(notification.Type, language, notification.Metadata)
	userID := user.ID.Hex()
	url := "/profile/" + userID

	if pushTexts.Title != "" {
		slog.Info("Sending push notification", "title", pushTexts.Title, "userId", userID)

		subsCtx, subsCancel := context.WithTimeout(context.Background(), 10*time.Second)
		subs, err := q.mongoDB.GetPushSubscriptionsForUsers(subsCtx, []string{userID})
		subsCancel()
		if err != nil {
			slog.Error("Error fetching push subscriptions for user", "userId", userID, "error", err)
		} else {
			slog.Info("Found push subscriptions for user", "count", len(subs), "userId", userID)
			for _, sub := range subs {
				go func(s models.PushSubscription) {
					if err := q.pushSender.SendNotification(s, pushTexts.Title, pushTexts.Message, url); err != nil {
						slog.Error("Error sending push notification", "subscriptionId", s.ID.Hex(), "error", err)
					} else {
						slog.Info("Push sent to subscription", "subscriptionId", s.ID.Hex())
					}
				}(sub)
			}
		}
	} else {
		slog.Warn("No push notification title set for type", "type", notification.Type)
	}

	return success
}

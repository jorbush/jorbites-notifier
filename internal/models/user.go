package models

import "go.mongodb.org/mongo-driver/v2/bson"

type NotificationPreferences struct {
	Social              *bool `bson:"social,omitempty"`
	NewContent          *bool `bson:"newContent,omitempty"`
	EventsAndChallenges *bool `bson:"eventsAndChallenges,omitempty"`
	Quests              *bool `bson:"quests,omitempty"`
	Voting              *bool `bson:"voting,omitempty"`
	Achievements        *bool `bson:"achievements,omitempty"`
}

type User struct {
	ID                      bson.ObjectID            `bson:"_id"`
	Email                   string                   `bson:"email"`
	EmailNotifications      bool                     `bson:"emailNotifications"`
	Language                *string                  `bson:"language,omitempty"`
	NotificationPreferences *NotificationPreferences `bson:"notificationPreferences,omitempty"`
}

func (u *User) IsNotificationCategoryEnabled(notifType NotificationType) bool {
	if notifType == TypeForgotPassword || notifType == TypeNotificationsActivated {
		return true
	}

	if u == nil || u.NotificationPreferences == nil {
		return true
	}

	prefs := u.NotificationPreferences

	switch notifType {
	case TypeNewLike, TypeNewComment, TypeMentionInComment:
		if prefs.Social != nil {
			return *prefs.Social
		}
	case TypeNewRecipe, TypeNewBlog:
		if prefs.NewContent != nil {
			return *prefs.NewContent
		}
	case TypeNewEvent, TypeEventEndingSoon, TypeNewChallenge:
		if prefs.EventsAndChallenges != nil {
			return *prefs.EventsAndChallenges
		}
	case TypeNewQuest, TypeQuestFulfilled:
		if prefs.Quests != nil {
			return *prefs.Quests
		}
	case TypeNewVotation, TypeVotationResult:
		if prefs.Voting != nil {
			return *prefs.Voting
		}
	case TypeNewBadge, TypeVerified:
		if prefs.Achievements != nil {
			return *prefs.Achievements
		}
	}

	return true
}

package models

import (
	"testing"
)

func boolPtr(b bool) *bool {
	return &b
}

func TestIsNotificationCategoryEnabled(t *testing.T) {
	tests := []struct {
		name      string
		user      *User
		notifType NotificationType
		expected  bool
	}{
		{
			name:      "nil user defaults all to true",
			user:      nil,
			notifType: TypeNewLike,
			expected:  true,
		},
		{
			name:      "user with nil preferences defaults all to true",
			user:      &User{NotificationPreferences: nil},
			notifType: TypeNewComment,
			expected:  true,
		},
		{
			name: "user with empty preferences defaults missing fields to true",
			user: &User{
				NotificationPreferences: &NotificationPreferences{},
			},
			notifType: TypeNewRecipe,
			expected:  true,
		},
		{
			name: "social disabled skips NEW_LIKE",
			user: &User{
				NotificationPreferences: &NotificationPreferences{
					Social: boolPtr(false),
				},
			},
			notifType: TypeNewLike,
			expected:  false,
		},
		{
			name: "social disabled skips NEW_COMMENT",
			user: &User{
				NotificationPreferences: &NotificationPreferences{
					Social: boolPtr(false),
				},
			},
			notifType: TypeNewComment,
			expected:  false,
		},
		{
			name: "social disabled skips MENTION_IN_COMMENT",
			user: &User{
				NotificationPreferences: &NotificationPreferences{
					Social: boolPtr(false),
				},
			},
			notifType: TypeMentionInComment,
			expected:  false,
		},
		{
			name: "social enabled allows NEW_LIKE",
			user: &User{
				NotificationPreferences: &NotificationPreferences{
					Social: boolPtr(true),
				},
			},
			notifType: TypeNewLike,
			expected:  true,
		},
		{
			name: "newContent disabled skips NEW_RECIPE and NEW_BLOG",
			user: &User{
				NotificationPreferences: &NotificationPreferences{
					NewContent: boolPtr(false),
				},
			},
			notifType: TypeNewRecipe,
			expected:  false,
		},
		{
			name: "newContent disabled skips NEW_BLOG",
			user: &User{
				NotificationPreferences: &NotificationPreferences{
					NewContent: boolPtr(false),
				},
			},
			notifType: TypeNewBlog,
			expected:  false,
		},
		{
			name: "eventsAndChallenges disabled skips NEW_EVENT",
			user: &User{
				NotificationPreferences: &NotificationPreferences{
					EventsAndChallenges: boolPtr(false),
				},
			},
			notifType: TypeNewEvent,
			expected:  false,
		},
		{
			name: "quests disabled skips NEW_QUEST",
			user: &User{
				NotificationPreferences: &NotificationPreferences{
					Quests: boolPtr(false),
				},
			},
			notifType: TypeNewQuest,
			expected:  false,
		},
		{
			name: "voting disabled skips NEW_VOTATION",
			user: &User{
				NotificationPreferences: &NotificationPreferences{
					Voting: boolPtr(false),
				},
			},
			notifType: TypeNewVotation,
			expected:  false,
		},
		{
			name: "achievements disabled skips NEW_BADGE",
			user: &User{
				NotificationPreferences: &NotificationPreferences{
					Achievements: boolPtr(false),
				},
			},
			notifType: TypeNewBadge,
			expected:  false,
		},
		{
			name: "FORGOT_PASSWORD always returns true even when everything disabled",
			user: &User{
				NotificationPreferences: &NotificationPreferences{
					Social:              boolPtr(false),
					NewContent:          boolPtr(false),
					EventsAndChallenges: boolPtr(false),
					Quests:              boolPtr(false),
					Voting:              boolPtr(false),
					Achievements:        boolPtr(false),
				},
			},
			notifType: TypeForgotPassword,
			expected:  true,
		},
		{
			name: "NOTIFICATIONS_ACTIVATED always returns true even when everything disabled",
			user: &User{
				NotificationPreferences: &NotificationPreferences{
					Social: boolPtr(false),
				},
			},
			notifType: TypeNotificationsActivated,
			expected:  true,
		},
		{
			name: "mixed preferences work correctly",
			user: &User{
				NotificationPreferences: &NotificationPreferences{
					Social:     boolPtr(false),
					NewContent: boolPtr(true),
				},
			},
			notifType: TypeNewLike,
			expected:  false,
		},
		{
			name: "mixed preferences allows enabled category",
			user: &User{
				NotificationPreferences: &NotificationPreferences{
					Social:     boolPtr(false),
					NewContent: boolPtr(true),
				},
			},
			notifType: TypeNewRecipe,
			expected:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.user.IsNotificationCategoryEnabled(tt.notifType)
			if result != tt.expected {
				t.Errorf("IsNotificationCategoryEnabled(%s) = %v, want %v", tt.notifType, result, tt.expected)
			}
		})
	}
}

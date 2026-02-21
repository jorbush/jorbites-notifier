package i18n

import (
	"testing"

	"github.com/jorbush/jorbites-notifier/internal/models"
)

func TestGetUserLanguage(t *testing.T) {
	tests := []struct {
		name     string
		user     *models.User
		expected string
	}{
		{
			name: "User with Spanish language",
			user: &models.User{
				Language: stringPtr("es"),
			},
			expected: "es",
		},
		{
			name: "User with Catalan language",
			user: &models.User{
				Language: stringPtr("ca"),
			},
			expected: "ca",
		},
		{
			name: "User with English language",
			user: &models.User{
				Language: stringPtr("en"),
			},
			expected: "en",
		},
		{
			name: "User with nil language (should default to es)",
			user: &models.User{
				Language: nil,
			},
			expected: "es",
		},
		{
			name: "User with empty language (should default to es)",
			user: &models.User{
				Language: stringPtr(""),
			},
			expected: "es",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetUserLanguage(tt.user)
			if result != tt.expected {
				t.Errorf("GetUserLanguage() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestGetEmailTemplateContent(t *testing.T) {
	notificationTypes := []models.NotificationType{
		models.TypeNewComment,
		models.TypeNewLike,
		models.TypeNewRecipe,
		models.TypeNotificationsActivated,
		models.TypeForgotPassword,
		models.TypeMentionInComment,
		models.TypeNewBlog,
		models.TypeNewQuest,
		models.TypeQuestFulfilled,
	}

	languages := []string{"es", "ca", "en"}

	for _, notifType := range notificationTypes {
		for _, lang := range languages {
			t.Run(string(notifType)+"_"+lang, func(t *testing.T) {
				content := GetEmailTemplateContent(notifType, lang)
				if content == "" {
					t.Errorf("GetEmailTemplateContent(%v, %s) returned empty string", notifType, lang)
				}
			})
		}
	}
}

func TestGetEmailTemplateContentFallback(t *testing.T) {
	// Test that unsupported language falls back to Spanish
	content := GetEmailTemplateContent(models.TypeNewComment, "fr")
	if content == "" {
		t.Error("GetEmailTemplateContent should fallback to Spanish for unsupported language")
	}

	// Test that unknown notification type returns empty string
	content = GetEmailTemplateContent("UNKNOWN_TYPE", "es")
	if content != "" {
		t.Error("GetEmailTemplateContent should return empty string for unknown notification type")
	}
}

func TestGetEmailSubject(t *testing.T) {
	notificationTypes := []models.NotificationType{
		models.TypeNewComment,
		models.TypeNewLike,
		models.TypeNewRecipe,
		models.TypeNotificationsActivated,
		models.TypeForgotPassword,
		models.TypeMentionInComment,
		models.TypeNewBlog,
		models.TypeNewQuest,
		models.TypeQuestFulfilled,
	}

	languages := []string{"es", "ca", "en"}

	for _, notifType := range notificationTypes {
		for _, lang := range languages {
			t.Run(string(notifType)+"_"+lang, func(t *testing.T) {
				subject := GetEmailSubject(notifType, lang)
				if subject == "" {
					t.Errorf("GetEmailSubject(%v, %s) returned empty string", notifType, lang)
				}
			})
		}
	}
}

func TestGetEmailSubjectFallback(t *testing.T) {
	tests := []struct {
		name         string
		notifType    models.NotificationType
		lang         string
		shouldContain string
	}{
		{
			name:         "Unsupported language fallback to Spanish",
			notifType:    models.TypeNewComment,
			lang:         "fr",
			shouldContain: "Comentario",
		},
		{
			name:         "Unknown notification type returns default",
			notifType:    "UNKNOWN_TYPE",
			lang:         "es",
			shouldContain: "Notificación",
		},
		{
			name:         "Unknown notification type with Catalan returns Catalan default",
			notifType:    "UNKNOWN_TYPE",
			lang:         "ca",
			shouldContain: "Notificació",
		},
		{
			name:         "Unknown notification type with English returns English default",
			notifType:    "UNKNOWN_TYPE",
			lang:         "en",
			shouldContain: "Notification",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			subject := GetEmailSubject(tt.notifType, tt.lang)
			if subject == "" {
				t.Errorf("GetEmailSubject() returned empty string")
			}
		})
	}
}

func TestGetBaseTemplateFooter(t *testing.T) {
	languages := []string{"es", "ca", "en"}

	for _, lang := range languages {
		t.Run(lang, func(t *testing.T) {
			footer := GetBaseTemplateFooter(lang)
			if footer == "" {
				t.Errorf("GetBaseTemplateFooter(%s) returned empty string", lang)
			}
		})
	}
}

func TestGetBaseTemplateFooterFallback(t *testing.T) {
	// Test unsupported language falls back to Spanish
	footer := GetBaseTemplateFooter("fr")
	if footer == "" {
		t.Error("GetBaseTemplateFooter should fallback to Spanish for unsupported language")
	}
}

func TestGetPushNotificationText(t *testing.T) {
	tests := []struct {
		name      string
		notifType models.NotificationType
		lang      string
		metadata  map[string]string
		wantTitle bool
		wantMsg   bool
	}{
		{
			name:      "TypeNewLike - Spanish",
			notifType: models.TypeNewLike,
			lang:      "es",
			metadata:  map[string]string{"likedBy": "John"},
			wantTitle: true,
			wantMsg:   true,
		},
		{
			name:      "TypeNewLike - Catalan",
			notifType: models.TypeNewLike,
			lang:      "ca",
			metadata:  map[string]string{"likedBy": "John"},
			wantTitle: true,
			wantMsg:   true,
		},
		{
			name:      "TypeNewLike - English",
			notifType: models.TypeNewLike,
			lang:      "en",
			metadata:  map[string]string{"likedBy": "John"},
			wantTitle: true,
			wantMsg:   true,
		},
		{
			name:      "TypeNewComment - Spanish",
			notifType: models.TypeNewComment,
			lang:      "es",
			metadata:  map[string]string{"authorName": "Jane"},
			wantTitle: true,
			wantMsg:   true,
		},
		{
			name:      "TypeNewComment - Catalan",
			notifType: models.TypeNewComment,
			lang:      "ca",
			metadata:  map[string]string{"authorName": "Jane"},
			wantTitle: true,
			wantMsg:   true,
		},
		{
			name:      "TypeNewComment - English",
			notifType: models.TypeNewComment,
			lang:      "en",
			metadata:  map[string]string{"authorName": "Jane"},
			wantTitle: true,
			wantMsg:   true,
		},
		{
			name:      "TypeNotificationsActivated - Spanish",
			notifType: models.TypeNotificationsActivated,
			lang:      "es",
			metadata:  map[string]string{},
			wantTitle: true,
			wantMsg:   true,
		},
		{
			name:      "TypeMentionInComment - Spanish",
			notifType: models.TypeMentionInComment,
			lang:      "es",
			metadata:  map[string]string{},
			wantTitle: true,
			wantMsg:   true,
		},
		{
			name:      "TypeNewRecipe - Spanish",
			notifType: models.TypeNewRecipe,
			lang:      "es",
			metadata:  map[string]string{"recipeName": "Paella"},
			wantTitle: true,
			wantMsg:   true,
		},
		{
			name:      "TypeNewBlog - Spanish",
			notifType: models.TypeNewBlog,
			lang:      "es",
			metadata:  map[string]string{"title": "New Post"},
			wantTitle: true,
			wantMsg:   true,
		},
		{
			name:      "TypeNewQuest - Spanish",
			notifType: models.TypeNewQuest,
			lang:      "es",
			metadata:  map[string]string{"questId": "quest-123"},
			wantTitle: true,
			wantMsg:   true,
		},
		{
			name:      "TypeNewQuest - Catalan",
			notifType: models.TypeNewQuest,
			lang:      "ca",
			metadata:  map[string]string{"questId": "quest-123"},
			wantTitle: true,
			wantMsg:   true,
		},
		{
			name:      "TypeNewQuest - English",
			notifType: models.TypeNewQuest,
			lang:      "en",
			metadata:  map[string]string{"questId": "quest-123"},
			wantTitle: true,
			wantMsg:   true,
		},
		{
			name:      "TypeQuestFulfilled - Spanish",
			notifType: models.TypeQuestFulfilled,
			lang:      "es",
			metadata:  map[string]string{"questId": "quest-123", "fulfilledByName": "User1"},
			wantTitle: true,
			wantMsg:   true,
		},
		{
			name:      "TypeQuestFulfilled - Catalan",
			notifType: models.TypeQuestFulfilled,
			lang:      "ca",
			metadata:  map[string]string{"questId": "quest-123", "fulfilledByName": "User1"},
			wantTitle: true,
			wantMsg:   true,
		},
		{
			name:      "TypeQuestFulfilled - English",
			notifType: models.TypeQuestFulfilled,
			lang:      "en",
			metadata:  map[string]string{"questId": "quest-123", "fulfilledByName": "User1"},
			wantTitle: true,
			wantMsg:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetPushNotificationText(tt.notifType, tt.lang, tt.metadata)
			
			if tt.wantTitle && result.Title == "" {
				t.Errorf("GetPushNotificationText() Title is empty, want non-empty")
			}
			if tt.wantMsg && result.Message == "" {
				t.Errorf("GetPushNotificationText() Message is empty, want non-empty")
			}
		})
	}
}

func TestGetPushNotificationTextWithMetadata(t *testing.T) {
	// Test that metadata is properly included in messages
	t.Run("NewLike includes likedBy name", func(t *testing.T) {
		metadata := map[string]string{"likedBy": "John Doe"}
		
		// Test Spanish
		result := GetPushNotificationText(models.TypeNewLike, "es", metadata)
		if result.Message != "John Doe le ha dado like a tu receta" {
			t.Errorf("Spanish message doesn't include likedBy name: %s", result.Message)
		}

		// Test Catalan
		result = GetPushNotificationText(models.TypeNewLike, "ca", metadata)
		if result.Message != "John Doe ha fet like a la teva recepta" {
			t.Errorf("Catalan message doesn't include likedBy name: %s", result.Message)
		}

		// Test English
		result = GetPushNotificationText(models.TypeNewLike, "en", metadata)
		if result.Message != "John Doe liked your recipe" {
			t.Errorf("English message doesn't include likedBy name: %s", result.Message)
		}
	})

	t.Run("NewComment includes author name", func(t *testing.T) {
		metadata := map[string]string{"authorName": "Jane Smith"}
		
		// Test Spanish
		result := GetPushNotificationText(models.TypeNewComment, "es", metadata)
		if result.Message != "Jane Smith ha comentado en tu receta" {
			t.Errorf("Spanish message doesn't include author name: %s", result.Message)
		}
	})

	t.Run("NewRecipe includes recipe name", func(t *testing.T) {
		metadata := map[string]string{"recipeName": "Tortilla de Patatas"}
		
		// Test Spanish
		result := GetPushNotificationText(models.TypeNewRecipe, "es", metadata)
		if result.Message != "Nueva receta disponible: Tortilla de Patatas" {
			t.Errorf("Spanish message doesn't include recipe name: %s", result.Message)
		}
	})
}

func TestGetPushNotificationTextFallback(t *testing.T) {
	// Test unsupported language falls back to Spanish
	result := GetPushNotificationText(models.TypeNewLike, "fr", map[string]string{"likedBy": "John"})
	if result.Title == "" || result.Message == "" {
		t.Error("GetPushNotificationText should return valid values for unsupported language")
	}
	// Should be Spanish
	if result.Title != "Nuevo Like" {
		t.Errorf("Expected Spanish fallback, got title: %s", result.Title)
	}

	// Test unknown notification type returns generic message
	result = GetPushNotificationText("UNKNOWN_TYPE", "es", map[string]string{})
	if result.Title == "" || result.Message == "" {
		t.Error("GetPushNotificationText should return default values for unknown notification type")
	}
}

func TestAllNotificationTypesHaveTranslations(t *testing.T) {
	notificationTypes := []models.NotificationType{
		models.TypeNewComment,
		models.TypeNewLike,
		models.TypeNewRecipe,
		models.TypeNotificationsActivated,
		models.TypeForgotPassword,
		models.TypeMentionInComment,
		models.TypeNewBlog,
		models.TypeNewQuest,
		models.TypeQuestFulfilled,
	}

	languages := []string{"es", "ca", "en"}

	for _, notifType := range notificationTypes {
		for _, lang := range languages {
			t.Run(string(notifType)+"_"+lang+"_email_content", func(t *testing.T) {
				content := GetEmailTemplateContent(notifType, lang)
				if content == "" {
					t.Errorf("Missing email template content for %s in %s", notifType, lang)
				}
			})

			t.Run(string(notifType)+"_"+lang+"_email_subject", func(t *testing.T) {
				subject := GetEmailSubject(notifType, lang)
				if subject == "" {
					t.Errorf("Missing email subject for %s in %s", notifType, lang)
				}
			})

			t.Run(string(notifType)+"_"+lang+"_push", func(t *testing.T) {
				// Only test push notifications for types that use them
				if notifType == models.TypeForgotPassword {
					return // Skip - forgot password doesn't send push
				}
				
				result := GetPushNotificationText(notifType, lang, map[string]string{
					"likedBy":         "Test User",
					"authorName":      "Test Author",
					"recipeName":      "Test Recipe",
					"title":           "Test Title",
					"fulfilledByName": "Test User",
				})
				if result.Title == "" || result.Message == "" {
					t.Errorf("Missing push notification for %s in %s", notifType, lang)
				}
			})
		}
	}
}

// Helper function for tests
func stringPtr(s string) *string {
	return &s
}

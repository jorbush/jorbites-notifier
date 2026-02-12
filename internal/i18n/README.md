# i18n - Internationalization Package

This package provides internationalization support for the Jorbites notification system.

## Supported Languages

- **es** - Spanish (default)
- **ca** - Catalan
- **en** - English

## Features

- Email template translations (subject, content, footer)
- Push notification translations (title, message)
- Automatic fallback to Spanish for missing translations
- Safe null handling for user language preferences

## Testing

The package includes comprehensive tests to ensure all notification types have translations in all supported languages.

### Run Tests

```bash
# Run all i18n tests
make test-i18n

# Run all project tests
make test

# Generate coverage report
make test-coverage
```

### Test Coverage

The test suite verifies:
- ✅ All notification types have translations in all languages
- ✅ Email templates (content and subjects) exist for all types
- ✅ Push notification texts exist for all types
- ✅ Fallback logic works correctly
- ✅ Metadata is properly included in messages
- ✅ Null/empty language handling

## Usage

```go
import "github.com/jorbush/jorbites-notifier/internal/i18n"

// Get user's preferred language (defaults to "es" if nil/empty)
language := i18n.GetUserLanguage(user)

// Get email template content
content := i18n.GetEmailTemplateContent(models.TypeNewComment, language)

// Get email subject
subject := i18n.GetEmailSubject(models.TypeNewComment, language)

// Get push notification text
pushTexts := i18n.GetPushNotificationText(
    models.TypeNewLike, 
    language, 
    map[string]string{"likedBy": "John Doe"},
)
```

## Adding New Languages

To add a new language:

1. Add translations to `emailTemplateContent` map
2. Add translations to `emailSubjects` map
3. Add translations to `baseTemplateFooter` map
4. Add cases to `GetPushNotificationText` function
5. Run tests to verify all translations are present

## Adding New Notification Types

When adding a new notification type:

1. Define the type in `internal/models/notification.go`
2. Add translations in all 3 languages to:
   - `emailTemplateContent`
   - `emailSubjects`
   - Push notification logic in `GetPushNotificationText`
3. Run `make test-i18n` to verify completeness

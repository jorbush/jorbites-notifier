# Email Service

## Overview

The Email Service is a core component of the Jorbites Notifier microservice, handling the delivery of various notification types to users. It provides an interface to send templated HTML emails with consistent styling while maintaining compatibility across different email clients.

## Architecture

The email functionality is implemented as a separate package (`internal/email`) to maintain a clean separation of concerns within the notification service. This modular approach allows for better maintainability and testability.

### Key Components

1. **EmailSender**: Manages the SMTP connection and handles the actual email delivery.
2. **Email Templates**: HTML templates for different notification types with a consistent layout.
3. **Template Rendering**: Dynamic content generation based on notification metadata.

## Implementation Details

### Directory Structure

```
internal/
├── email/
     ├── sender.go     # Email sending functionality
     └── templates.go  # Email templates and
```

### Email Templates

Email templates are HTML-based with a responsive design that works across desktop and mobile email clients. All emails follow a consistent layout:

- Header with Jorbites logo
- Content section with notification-specific message
- Footer with standard information and links

Templates use Go's built-in `text/template` package for variable substitution and dynamic content generation.

### Supported Notification Types

1. **NEW_COMMENT**: Sent when a user comments on a recipe
2. **NEW_LIKE**: Sent when a user likes a recipe
3. **NEW_RECIPE**: Sent when a new recipe is published
4. **NOTIFICATIONS_ACTIVED**: Sent when a user activates notifications

Each notification type has its own subject line and body content, while maintaining the consistent header and footer.

## Configuration

The email service uses SMTP for sending emails, configured through environment variables:

| Variable | Description | Default |
|----------|-------------|---------|
| `SMTP_HOST` | SMTP server hostname | smtp.gmail.com |
| `SMTP_PORT` | SMTP server port | 587 |
| `SMTP_USER` | SMTP authentication username | *Required* |
| `SMTP_PASSWORD` | SMTP authentication password | *Required* |

### Example .env file

```
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=notifications@jorbites.com
SMTP_PASSWORD=your-secure-password
API_KEY=your-api-key
```

## Usage Examples

### Sending a Notification Email

To send an email notification, create a POST request to the `/notifications` endpoint:

```bash
curl -X POST http://localhost:8080/notifications \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-api-key" \
  -d '{
    "type": "NEW_COMMENT",
    "recipient": "user@example.com",
    "metadata": {
      "authorName": "John Doe",
      "recipeId": "12345"
    }
  }'
```

### Internal Implementation

```go
// Create and send a notification
notification := models.Notification{
    Type:      models.TypeNewComment,
    Recipient: "user@example.com",
    Metadata: map[string]string{
        "authorName": "John Doe",
        "recipeId":   "12345",
    },
}

// The notification is added to the queue
queue.Enqueue(notification)

// When processed, EmailSender handles delivery
emailSender.SendNotificationEmail(notification)
```

## Implementation Notes

### SMTP Client

The implementation uses Go's native `net/smtp` package to avoid external dependencies. This provides:

- Basic SMTP authentication
- TLS support
- Header construction

### Email Headers

Each email is constructed with proper MIME headers:

```
From: Jorbites <notifications@jorbites.com>
Subject: New Comment on Your Recipe - Jorbites
To: user@example.com
MIME-version: 1.0
Content-Type: text/html; charset="UTF-8"
```

The `From` header includes a display name ("Jorbites") to improve the recipient's inbox experience.

### Logo Handling

For maximum compatibility across email clients:

1. The logo is hosted on a public URL rather than embedded as base64
2. The URL is absolute to ensure accessibility regardless of email client

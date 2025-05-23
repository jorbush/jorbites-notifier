# Notification Types

Jorbites Notifier supports several types of notifications. Each type triggers a specific behavior when processed.

## Supported Notification Types

### NEW_COMMENT

Sent when a user comments on a recipe.

**Metadata Fields**:
- `commentId`: ID of the new comment
- `authorName`: Name of the comment author
- `recipeId`: ID of the recipe that was commented on

**Example**:
```json
{
  "type": "NEW_COMMENT",
  "recipient": "user@example.com",
  "metadata": {
    "commentId": "12345",
    "authorName": "User1",
    "recipeId": "67890"
  }
}
```

### NEW_LIKE

Sent when a user likes a recipe.

**Metadata Fields**:
- `likedBy`: Name of the user who liked the recipe
- `recipeId`: ID of the recipe that was liked

**Example**:
```json
{
  "type": "NEW_LIKE",
  "recipient": "user@example.com",
  "metadata": {
    "likedBy": "User2",
    "recipeId": "67890"
  }
}
```

### NEW_RECIPE

Sent when a user publishes a new recipe.

**Metadata Fields**:
- `recipeId`: ID of the new recipe

**Example**:
```json
{
  "type": "NEW_RECIPE",
  "recipient": "user@example.com",
  "metadata": {
    "recipeId": "67890",
  }
}
```

### NOTIFICATIONS_ACTIVATED

Sent when a user activates notifications for their account.

**Example**:
```json
{
  "type": "NOTIFICATIONS_ACTIVATED",
  "recipient": "user@example.com"
}
```

### FORGOT_PASSWORD

Sent when a user requests a password reset.

**Example**:
```json
{
  "type": "FORGOT_PASSWORD",
  "recipient": "user@example.com",
  "metadata": {
    "resetUrl": "https://example.com/reset-password?token=abc123"
  }
}
```

## Adding New Notification Types

To add a new notification type:

1. Add the type constant in `models/notification.go`
2. Implement the processing logic for the new type in `queue.processNotificationByType()`
3. Update this documentation with details about the new type

# API Reference

Jorbites Notifier exposes a RESTful API for managing notifications.

## Endpoints

### Health Check

```
GET /health
```

Checks if the service is up and running.

#### Response

```json
{
  "status": "ok"
}
```

### Enqueue Notification

```
POST /notifications
```

Adds a new notification to the processing queue.

#### Request Body

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

#### Fields

| Field | Type | Description | Required |
|-------|------|-------------|----------|
| `type` | string | Type of notification (see [Notification Types](./notification-types.md)) | Yes |
| `recipient` | string | Email address of the recipient | No |
| `metadata` | object | Additional data needed for the notification | No |

#### Response

Returns the created notification object with status code 201 (Created):

```json
{
  "id": "f47ac10b-58cc-4372-a567-0e02b2c3d479",
  "type": "NEW_COMMENT",
  "status": "pending",
  "recipient": "user@example.com",
  "metadata": {
    "commentId": "12345",
    "authorName": "User1",
    "recipeId": "67890"
  }
}
```

### Get Queue Status

```
GET /queue
```

Returns the current status of the notification queue.

#### Response

```json
{
  "count": 2,
  "notifications": [
    {
      "id": "f47ac10b-58cc-4372-a567-0e02b2c3d479",
      "type": "NEW_COMMENT",
      "status": "processing",
      "recipient": "user@example.com",
      "metadata": {
        "commentId": "12345",
        "authorName": "User1",
        "recipeId": "67890"
      }
    },
    {
      "id": "a1b2c3d4-e5f6-4a5b-8c7d-9e0f1a2b3c4d",
      "type": "NEW_LIKE",
      "status": "pending",
      "recipient": "user2@example.com",
      "metadata": {
        "likedBy": "User2",
        "recipeId": "67890"
      }
    }
  ]
}
```

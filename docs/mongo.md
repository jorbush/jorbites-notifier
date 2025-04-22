# MongoDB Integration

## Overview

This documentation covers the integration of MongoDB with the Jorbites Notifier service, specifically for handling broadcast notifications such as the `NEW_RECIPE` notification type. This enhancement allows the notifier to directly access user data from MongoDB and broadcast notifications to all users who have enabled email notifications.

## Architecture Changes

The notifier service now connects to MongoDB to fetch user data directly, allowing it to handle broadcast notifications without requiring the client to specify individual recipients. This significantly improves scalability by:

1. Reducing API traffic between services
2. Eliminating the need for the client to fetch and process users
3. Centralizing notification logic in the dedicated notification service

## MongoDB Configuration

### Environment Variables

Add the following environment variables to your `.env` file:

```
MONGO_URI={your_mongo_uri}
MONGO_DB={your_database_name}
```

### Configuration Structure

The MongoDB connection details have been added to the configuration structure:

```go
type Config struct {
	Port         string
	SMTPHost     string
	SMTPPort     int
	SMTPUser     string
	SMTPPassword string
	MongoURI     string
	MongoDB      string
}
```

## MongoDB Client Implementation

### MongoDB Connection

```go
// NewMongoDB creates a new MongoDB client instance
func NewMongoDB(cfg *config.Config) (*MongoDB, error) {
	clientOptions := options.Client().ApplyURI(cfg.MongoURI)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, err
	}
	log.Println("Connected to MongoDB successfully")
	db := client.Database(cfg.MongoDB)
	return &MongoDB{
		client: client,
		db:     db,
	}, nil
}
```

### User Data Access

```go
// GetUsersWithNotificationsEnabled returns all users who have email notifications enabled
func (m *MongoDB) GetUsersWithNotificationsEnabled(ctx context.Context) ([]models.User, error) {
	collection := m.db.Collection("users")

	filter := bson.D{{"emailNotifications", true}}
	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var users []models.User
	if err = cursor.All(ctx, &users); err != nil {
		return nil, err
	}

	log.Printf("Found %d users with email notifications enabled", len(users))
	return users, nil
}
```

## Queue Modifications for Broadcast Notifications

### Queue Structure Update

The queue now includes a MongoDB client:

```go
type Queue struct {
	notifications []models.Notification
	mutex         sync.Mutex
	processing    bool
	notifyChan    chan struct{}
	emailSender   *email.EmailSender
	mongoDB       *database.MongoDB
}
```

### Queue Initialization

Queue initialization now includes connecting to MongoDB:

```go
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
		mongoDB:       mongoDB,
	}
}
```

### NEW_RECIPE Notification Processing

A dedicated method handles the `NEW_RECIPE` notification type by fetching users from MongoDB and sending individual emails:

```go
func (q *Queue) processNewRecipeNotification(notification models.Notification) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Get all users with email notifications enabled
	users, err := q.mongoDB.GetUsersWithNotificationsEnabled(ctx)
	if err != nil {
		log.Printf("Error fetching users for notification %s: %v", notification.ID, err)
		return false
	}

	log.Printf("Sending new recipe notification to %d users with notifications enabled", len(users))

	successCount := 0
	failCount := 0

	// Process each user in batches
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

		// Add a small delay to avoid overwhelming the SMTP server
		time.Sleep(100 * time.Millisecond)
	}

	log.Printf("New recipe notification results: %d successful, %d failed", successCount, failCount)
	return successCount > 0
}
```

### Notification Type Routing

The notification processing method now routes `NEW_RECIPE` notifications to the specialized handler:

```go
func (q *Queue) processNotificationByType(notification models.Notification) bool {
	switch notification.Type {
	case models.TypeNewRecipe:
		return q.processNewRecipeNotification(notification)
	case models.TypeNewComment, models.TypeNewLike, models.TypeNotificationsActivated:
		success, err := q.emailSender.SendNotificationEmail(notification)
		if err != nil {
			log.Printf("Error sending email for notification %s: %v", notification.ID, err)
			return false
		}
		return success
	default:
		log.Printf("Unknown notification type: %s", notification.Type)
		return false
	}
}
```

## API Usage for NEW_RECIPE Notifications

### Request Format

For `NEW_RECIPE` notifications, the client doesn't need to specify a recipient - the service will fetch all users with notifications enabled:

```json
{
  "type": "NEW_RECIPE",
  "metadata": {
    "recipeId": "12345",
    "recipeName": "Delicious Pasta Carbonara"
  }
}
```

### Example cURL Request

```bash
curl -X POST http://localhost:8080/notifications \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-api-key" \
  -d '{
    "type": "NEW_RECIPE",
    "metadata": {
      "recipeId": "12345",
      "recipeName": "Delicious Pasta Carbonara"
    }
  }'
```

## Performance Considerations

### Email Batch Processing

To prevent overwhelming your SMTP server, the implementation includes:

1. Sequential processing of user emails rather than parallel sending
2. A small delay (100ms) between email sends
3. Detailed success/failure tracking for monitoring

### MongoDB Query Optimization

The MongoDB query only fetches users with `emailNotifications: true`, ensuring we only process relevant users.

### Error Handling

The implementation includes robust error handling:

1. Individual email failures don't affect other users
2. MongoDB connection issues are logged and can be monitored
3. Processing timeouts are implemented to prevent hung operations

## Example Complete Flow

1. User publishes a new recipe in the NextJS application
2. NextJS sends a single `NEW_RECIPE` notification to the notifier service
3. Notifier service receives the notification and queues it
4. Queue processor detects the `NEW_RECIPE` type
5. Processor queries MongoDB for all users with notifications enabled
6. Processor creates individual email notifications for each user
7. Emails are sent sequentially with rate limiting
8. Results are logged with success/failure counts

This flow significantly reduces load on the NextJS application while ensuring all users receive timely notifications about new recipes.

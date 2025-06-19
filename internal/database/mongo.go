package database

import (
	"context"
	"log"
	"strings"

	"github.com/jorbush/jorbites-notifier/config"
	"github.com/jorbush/jorbites-notifier/internal/models"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type MongoDB struct {
	client *mongo.Client
	db     *mongo.Database
}

func NewMongoDB(cfg *config.Config) (*MongoDB, error) {
	clientOptions := options.Client().ApplyURI(cfg.MongoURI)
	client, err := mongo.Connect(clientOptions)
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

func (m *MongoDB) Close(ctx context.Context) error {
	return m.client.Disconnect(ctx)
}

// GetUsersWithNotificationsEnabled returns all users who have email notifications enabled
func (m *MongoDB) GetUsersWithNotificationsEnabled(ctx context.Context) ([]models.User, error) {
	collection := m.db.Collection("User")

	filter := bson.D{{Key: "emailNotifications", Value: true}}
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

func (m *MongoDB) GetUsersMentionedInComment(ctx context.Context, mentionedUsersIds string, recipientEmail string) ([]models.User, error) {
	collection := m.db.Collection("User")
	mentionedUserIdsArray := strings.Split(mentionedUsersIds, ",")
	var objectIds []bson.ObjectID
	for _, idStr := range mentionedUserIdsArray {
		objectId, err := bson.ObjectIDFromHex(idStr)
		if err != nil {
			log.Printf("Invalid ObjectID: %s, error: %v", idStr, err)
			continue // Skip invalid IDs
		}
		objectIds = append(objectIds, objectId)
	}
	if len(objectIds) == 0 {
		log.Printf("No valid ObjectIDs found")
		return []models.User{}, nil
	}
	filter := bson.D{
		{Key: "_id", Value: bson.D{{Key: "$in", Value: objectIds}}},
		{Key: "emailNotifications", Value: true},
		{Key: "email", Value: bson.D{{Key: "$ne", Value: recipientEmail}}},
	}

	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var users []models.User
	if err = cursor.All(ctx, &users); err != nil {
		return nil, err
	}

	log.Printf("Found %d users with email notifications enabled and mentioned in comment", len(users))
	return users, nil
}

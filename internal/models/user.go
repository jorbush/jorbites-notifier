package models

import "go.mongodb.org/mongo-driver/v2/bson"

type User struct {
	ID                 bson.ObjectID `bson:"_id"`
	Email              string        `bson:"email"`
	EmailNotifications bool          `bson:"emailNotifications"`
}

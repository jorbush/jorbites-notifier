package models

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type PushSubscription struct {
	ID        bson.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID    string        `bson:"userId" json:"userId"`
	Endpoint  string        `bson:"endpoint" json:"endpoint"`
	P256dh    string        `bson:"p256dh" json:"p256dh"`
	Auth      string        `bson:"auth" json:"auth"`
	CreatedAt time.Time     `bson:"createdAt" json:"createdAt"`
	UpdatedAt time.Time     `bson:"updatedAt" json:"updatedAt"`
}

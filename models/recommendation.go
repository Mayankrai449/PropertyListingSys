package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Recommendation struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	RecommenderID primitive.ObjectID `bson:"recommenderId" json:"recommenderId"`
	RecipientID   primitive.ObjectID `bson:"recipientId" json:"recipientId"`
	PropertyID    string             `bson:"propertyId" json:"propertyId"`
	CreatedAt     time.Time          `bson:"createdAt" json:"createdAt"`
}

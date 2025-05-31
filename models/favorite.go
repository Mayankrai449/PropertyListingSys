package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Favorite struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID     primitive.ObjectID `bson:"userId" json:"userId"`
	PropertyID string             `bson:"propertyId" json:"propertyId"`
	CreatedAt  time.Time          `bson:"createdAt" json:"createdAt"`
}

package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID        primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Email     string             `json:"email" bson:"email" validate:"required,email"`
	Password  string             `json:"password,omitempty" bson:"password" validate:"required,min=6"`
	Name      string             `json:"name" bson:"name" validate:"required"`
	Phone     string             `json:"phone,omitempty" bson:"phone"`
	Role      string             `json:"role" bson:"role" default:"user"`
	IsActive  bool               `json:"is_active" bson:"is_active" default:"true"`
	CreatedAt time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time          `json:"updated_at" bson:"updated_at"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type RegisterRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
	Name     string `json:"name" validate:"required"`
	Phone    string `json:"phone"`
}

type LoginResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

type UpdateUserRequest struct {
	Name  string `json:"name"`
	Phone string `json:"phone"`
}

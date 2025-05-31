package handlers

import (
	"PropertyListingSys/config"
	"PropertyListingSys/models"
	"PropertyListingSys/utils"
	"context"
	"net/http"
	"os"
	"time"

	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type UserController struct {
	collection *mongo.Collection
}

func NewUserController() *UserController {
	collectionName := os.Getenv("MONGODB_COLLECTION_USER")
	if collectionName == "" {
		collectionName = "user"
	}
	return &UserController{
		collection: config.GetCollection(collectionName),
	}
}

func (uc *UserController) Register(c echo.Context) error {
	var req models.RegisterRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request body",
		})
	}

	var existingUser models.User
	err := uc.collection.FindOne(context.Background(), bson.M{"email": req.Email}).Decode(&existingUser)
	if err == nil {
		return c.JSON(http.StatusConflict, map[string]string{
			"error": "User with this email already exists",
		})
	}

	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to hash password",
		})
	}

	user := models.User{
		ID:        primitive.NewObjectID(),
		Email:     req.Email,
		Password:  hashedPassword,
		Name:      req.Name,
		Phone:     req.Phone,
		Role:      "user",
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	_, err = uc.collection.InsertOne(context.Background(), user)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to create user",
		})
	}

	token, err := utils.GenerateJWT(user.ID, user.Email, user.Role)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to generate token",
		})
	}

	user.Password = ""

	return c.JSON(http.StatusCreated, models.LoginResponse{
		Token: token,
		User:  user,
	})
}

func (uc *UserController) Login(c echo.Context) error {
	var req models.LoginRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request body",
		})
	}

	var user models.User
	err := uc.collection.FindOne(context.Background(), bson.M{"email": req.Email}).Decode(&user)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "Invalid email or password",
		})
	}

	if !user.IsActive {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "Account is deactivated",
		})
	}

	err = utils.CheckPassword(user.Password, req.Password)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "Invalid email or password",
		})
	}

	token, err := utils.GenerateJWT(user.ID, user.Email, user.Role)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to generate token",
		})
	}

	user.Password = ""

	return c.JSON(http.StatusOK, models.LoginResponse{
		Token: token,
		User:  user,
	})
}

func (uc *UserController) GetProfile(c echo.Context) error {
	userID := c.Get("user_id").(primitive.ObjectID)

	var user models.User
	err := uc.collection.FindOne(context.Background(), bson.M{"_id": userID}).Decode(&user)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "User not found",
		})
	}

	user.Password = ""

	return c.JSON(http.StatusOK, user)
}

func (uc *UserController) UpdateProfile(c echo.Context) error {
	userID := c.Get("user_id").(primitive.ObjectID)

	var req models.UpdateUserRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request body",
		})
	}

	updateDoc := bson.M{
		"updated_at": time.Now(),
	}

	if req.Name != "" {
		updateDoc["name"] = req.Name
	}
	if req.Phone != "" {
		updateDoc["phone"] = req.Phone
	}

	_, err := uc.collection.UpdateOne(
		context.Background(),
		bson.M{"_id": userID},
		bson.M{"$set": updateDoc},
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to update user",
		})
	}

	var user models.User
	err = uc.collection.FindOne(context.Background(), bson.M{"_id": userID}).Decode(&user)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to fetch updated user",
		})
	}

	user.Password = ""

	return c.JSON(http.StatusOK, user)
}

func (uc *UserController) DeleteAccount(c echo.Context) error {
	userID := c.Get("user_id").(primitive.ObjectID)

	_, err := uc.collection.DeleteOne(context.Background(), bson.M{"_id": userID})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to delete user",
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Account deleted successfully",
	})
}

func (uc *UserController) GetAllUsers(c echo.Context) error {
	userRole := c.Get("user_role").(string)
	if userRole != "admin" {
		return c.JSON(http.StatusForbidden, map[string]string{
			"error": "Access denied",
		})
	}

	cursor, err := uc.collection.Find(context.Background(), bson.M{})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to fetch users",
		})
	}
	defer cursor.Close(context.Background())

	var users []models.User
	for cursor.Next(context.Background()) {
		var user models.User
		if err := cursor.Decode(&user); err != nil {
			continue
		}
		user.Password = ""
		users = append(users, user)
	}

	return c.JSON(http.StatusOK, users)
}

func (uc *UserController) SearchUserByEmail(c echo.Context) error {
	email := c.QueryParam("email")
	if email == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Email is required"})
	}
	var user models.User
	err := uc.collection.FindOne(context.Background(), bson.M{"email": email}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "User not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to search user"})
	}
	return c.JSON(http.StatusOK, map[string]string{"id": user.ID.Hex(), "name": user.Name})
}

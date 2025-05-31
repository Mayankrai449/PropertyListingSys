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

type FavoriteController struct {
	collection *mongo.Collection
}

func NewFavoriteController() *FavoriteController {
	collectionName := os.Getenv("MONGODB_COLLECTION_FAVORITES")
	if collectionName == "" {
		collectionName = "favorites"
	}
	return &FavoriteController{
		collection: config.GetCollection(collectionName),
	}
}

func (fc *FavoriteController) CreateFavorite(c echo.Context) error {
	userID := c.Get("user_id").(primitive.ObjectID)
	propertyID := c.FormValue("propertyId")
	if !utils.IsValidExternalID(propertyID) {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid property ID"})
	}
	count, err := fc.collection.CountDocuments(context.Background(), bson.M{"userId": userID, "propertyId": propertyID})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to check favorite"})
	}
	if count > 0 {
		return c.JSON(http.StatusConflict, map[string]string{"error": "Property already favorited"})
	}
	favorite := models.Favorite{
		ID:         primitive.NewObjectID(),
		UserID:     userID,
		PropertyID: propertyID,
		CreatedAt:  time.Now(),
	}
	_, err = fc.collection.InsertOne(context.Background(), favorite)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to favorite property"})
	}
	return c.JSON(http.StatusCreated, favorite)
}

func (fc *FavoriteController) GetFavorites(c echo.Context) error {
	userID := c.Get("user_id").(primitive.ObjectID)
	cursor, err := fc.collection.Find(context.Background(), bson.M{"userId": userID})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch favorites"})
	}
	defer cursor.Close(context.Background())
	var favorites []models.Favorite
	for cursor.Next(context.Background()) {
		var favorite models.Favorite
		if err := cursor.Decode(&favorite); err != nil {
			continue
		}
		favorites = append(favorites, favorite)
	}
	return c.JSON(http.StatusOK, favorites)
}

func (fc *FavoriteController) DeleteFavorite(c echo.Context) error {
	userID := c.Get("user_id").(primitive.ObjectID)
	propertyID := c.Param("propertyId")
	if !utils.IsValidExternalID(propertyID) {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid property ID"})
	}
	_, err := fc.collection.DeleteOne(context.Background(), bson.M{"userId": userID, "propertyId": propertyID})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to remove favorite"})
	}
	return c.JSON(http.StatusOK, map[string]string{"message": "Favorite removed successfully"})
}
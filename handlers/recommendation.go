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

type RecommendationController struct {
	collection     *mongo.Collection
	userCollection *mongo.Collection
}

func NewRecommendationController() *RecommendationController {
	collectionName := os.Getenv("MONGODB_COLLECTION_RECOMMENDATIONS")
	if collectionName == "" {
		collectionName = "recommendations"
	}
	userCollectionName := os.Getenv("MONGODB_COLLECTION_USER")
	if userCollectionName == "" {
		userCollectionName = "user"
	}
	return &RecommendationController{
		collection:     config.GetCollection(collectionName),
		userCollection: config.GetCollection(userCollectionName),
	}
}

func (rc *RecommendationController) CreateRecommendation(c echo.Context) error {
	recommenderID := c.Get("user_id").(primitive.ObjectID)
	var req struct {
		RecipientEmail string `json:"recipientEmail"`
		PropertyID     string `json:"propertyId"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
	}
	var recipient models.User
	err := rc.userCollection.FindOne(context.Background(), bson.M{"email": req.RecipientEmail}).Decode(&recipient)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Recipient not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to find recipient"})
	}
	if !utils.IsValidExternalID(req.PropertyID) {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid property ID"})
	}
	recommendation := models.Recommendation{
		ID:            primitive.NewObjectID(),
		RecommenderID: recommenderID,
		RecipientID:   recipient.ID,
		PropertyID:    req.PropertyID,
		CreatedAt:     time.Now(),
	}
	_, err = rc.collection.InsertOne(context.Background(), recommendation)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create recommendation"})
	}
	return c.JSON(http.StatusCreated, recommendation)
}

func (rc *RecommendationController) GetReceivedRecommendations(c echo.Context) error {
	userID := c.Get("user_id").(primitive.ObjectID)
	cursor, err := rc.collection.Find(context.Background(), bson.M{"recipientId": userID})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch recommendations"})
	}
	defer cursor.Close(context.Background())
	var recommendations []models.Recommendation
	for cursor.Next(context.Background()) {
		var rec models.Recommendation
		if err := cursor.Decode(&rec); err != nil {
			continue
		}
		recommendations = append(recommendations, rec)
	}
	return c.JSON(http.StatusOK, recommendations)
}

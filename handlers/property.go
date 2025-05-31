package handlers

import (
	"PropertyListingSys/config"
	"PropertyListingSys/models"
	"PropertyListingSys/utils"
	"context"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type PropertyController struct {
	collection *mongo.Collection
}

func NewPropertyController() *PropertyController {
	collectionName := os.Getenv("MONGODB_COLLECTION_PROPERTIES")
	if collectionName == "" {
		collectionName = "properties"
	}
	return &PropertyController{
		collection: config.GetCollection(collectionName),
	}
}

func (pc *PropertyController) CreateProperty(c echo.Context) error {
	userID := c.Get("user_id").(primitive.ObjectID)
	var property models.Property
	if err := c.Bind(&property); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
	}

	if !utils.IsValidExternalID(property.ExternalID) {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid externalId: must be PROP followed by a number greater than 1000"})
	}

	count, err := pc.collection.CountDocuments(context.Background(), bson.M{"_id": property.ExternalID})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to check property existence"})
	}
	if count > 0 {
		return c.JSON(http.StatusConflict, map[string]string{"error": "Property with this externalId already exists"})
	}

	property.CreatedBy = &userID
	property.CreatedAt = time.Now()
	property.UpdatedAt = time.Now()
	_, err = pc.collection.InsertOne(context.Background(), property)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create property"})
	}

	utils.RedisClient.Del(context.Background(), "properties:*")

	return c.JSON(http.StatusCreated, property)
}

func (pc *PropertyController) GetProperty(c echo.Context) error {
	id := c.Param("id")
	if !utils.IsValidExternalID(id) {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid property ID"})
	}

	var property models.Property
	cacheKey := "property:" + id
	ctx := context.Background()
	if hit, err := utils.GetCached(ctx, cacheKey, &property); hit && err == nil {
		return c.JSON(http.StatusOK, property)
	}

	err := pc.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&property)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Property not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch property"})
	}

	if err := utils.SetCached(ctx, cacheKey, property, 30*time.Second); err != nil {
	}

	return c.JSON(http.StatusOK, property)
}

func (pc *PropertyController) PatchProperty(c echo.Context) error {
	userID := c.Get("user_id").(primitive.ObjectID)
	userRole := c.Get("user_role").(string)
	id := c.Param("id")
	if !utils.IsValidExternalID(id) {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid property ID"})
	}

	var property models.Property
	err := pc.collection.FindOne(context.Background(), bson.M{"_id": id}).Decode(&property)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Property not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch property"})
	}

	if (property.CreatedBy != nil && *property.CreatedBy != userID && userRole != "admin") || (property.CreatedBy == nil && userRole != "admin") {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "You are not authorized to update this property"})
	}

	var update map[string]interface{}
	if err := c.Bind(&update); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
	}

	updateDoc := bson.M{"updatedAt": time.Now()}
	allowedFields := map[string]bool{
		"title":         true,
		"type":          true,
		"price":         true,
		"state":         true,
		"city":          true,
		"areaSqFt":      true,
		"bedrooms":      true,
		"bathrooms":     true,
		"amenities":     true,
		"furnished":     true,
		"availableFrom": true,
		"listedBy":      true,
		"tags":          true,
		"colorTheme":    true,
		"rating":        true,
		"isVerified":    true,
		"listingType":   true,
	}

	for key, value := range update {
		if allowedFields[key] {
			if key == "availableFrom" {
				if str, ok := value.(string); ok {
					if t, err := time.Parse(time.RFC3339, str); err == nil {
						updateDoc[key] = t
					} else {
						return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid availableFrom format"})
					}
				}
			} else {
				updateDoc[key] = value
			}
		}
	}

	if len(updateDoc) <= 1 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "No valid fields to update"})
	}

	_, err = pc.collection.UpdateOne(context.Background(), bson.M{"_id": id}, bson.M{"$set": updateDoc})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update property"})
	}

	err = pc.collection.FindOne(context.Background(), bson.M{"_id": id}).Decode(&property)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch updated property"})
	}

	cacheKey := "property:" + id
	utils.RedisClient.Del(context.Background(), cacheKey)
	utils.RedisClient.Del(context.Background(), "properties:*")

	return c.JSON(http.StatusOK, property)
}

func (pc *PropertyController) DeleteProperty(c echo.Context) error {
	userID := c.Get("user_id").(primitive.ObjectID)
	userRole := c.Get("user_role").(string)
	id := c.Param("id")
	if !utils.IsValidExternalID(id) {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid property ID"})
	}
	var property models.Property
	err := pc.collection.FindOne(context.Background(), bson.M{"_id": id}).Decode(&property)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Property not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch property"})
	}
	if (property.CreatedBy != nil && *property.CreatedBy != userID) || (property.CreatedBy == nil && userRole != "admin") {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "You are not authorized to delete this property"})
	}
	_, err = pc.collection.DeleteOne(context.Background(), bson.M{"_id": id})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete property"})
	}

	cacheKey := "property:" + id
	utils.RedisClient.Del(context.Background(), cacheKey)
	utils.RedisClient.Del(context.Background(), "properties:*")

	return c.JSON(http.StatusOK, map[string]string{"message": "Property deleted successfully"})
}

func (pc *PropertyController) ListProperties(c echo.Context) error {
	query := bson.M{}
	queryParams := make(map[string]string)

	if title := c.QueryParam("title"); title != "" {
		query["title"] = bson.M{"$regex": title, "$options": "i"}
		queryParams["title"] = title
	}
	if propType := c.QueryParam("type"); propType != "" {
		query["type"] = propType
		queryParams["type"] = propType
	}
	if priceMin := c.QueryParam("price_min"); priceMin != "" {
		if min, err := strconv.ParseFloat(priceMin, 64); err == nil {
			query["price"] = bson.M{"$gte": min}
			queryParams["price_min"] = priceMin
		}
	}
	if priceMax := c.QueryParam("price_max"); priceMax != "" {
		if max, err := strconv.ParseFloat(priceMax, 64); err == nil {
			if existing, ok := query["price"].(bson.M); ok {
				existing["$lte"] = max
			} else {
				query["price"] = bson.M{"$lte": max}
			}
			queryParams["price_max"] = priceMax
		}
	}
	if state := c.QueryParam("state"); state != "" {
		query["state"] = state
		queryParams["state"] = state
	}
	if city := c.QueryParam("city"); city != "" {
		query["city"] = city
		queryParams["city"] = city
	}
	if areaMin := c.QueryParam("area_min"); areaMin != "" {
		if min, err := strconv.ParseFloat(areaMin, 64); err == nil {
			query["areaSqFt"] = bson.M{"$gte": min}
			queryParams["area_min"] = areaMin
		}
	}
	if areaMax := c.QueryParam("area_max"); areaMax != "" {
		if max, err := strconv.ParseFloat(areaMax, 64); err == nil {
			if existing, ok := query["areaSqFt"].(bson.M); ok {
				existing["$lte"] = max
			} else {
				query["areaSqFt"] = bson.M{"$lte": max}
			}
			queryParams["area_max"] = areaMax
		}
	}
	if bedrooms := c.QueryParam("bedrooms"); bedrooms != "" {
		if num, err := strconv.Atoi(bedrooms); err == nil {
			query["bedrooms"] = num
			queryParams["bedrooms"] = bedrooms
		}
	}
	if bathrooms := c.QueryParam("bathrooms"); bathrooms != "" {
		if num, err := strconv.Atoi(bathrooms); err == nil {
			query["bathrooms"] = num
			queryParams["bathrooms"] = bathrooms
		}
	}
	if amenities := c.QueryParam("amenities"); amenities != "" {
		query["amenities"] = bson.M{"$regex": amenities, "$options": "i"}
		queryParams["amenities"] = amenities
	}
	if furnished := c.QueryParam("furnished"); furnished != "" {
		query["furnished"] = furnished
		queryParams["furnished"] = furnished
	}
	if availableFrom := c.QueryParam("available_from"); availableFrom != "" {
		if date, err := time.Parse("2006-01-02", availableFrom); err == nil {
			query["availableFrom"] = bson.M{"$gte": date}
			queryParams["available_from"] = availableFrom
		}
	}
	if listedBy := c.QueryParam("listed_by"); listedBy != "" {
		query["listedBy"] = listedBy
		queryParams["listed_by"] = listedBy
	}
	if tags := c.QueryParam("tags"); tags != "" {
		query["tags"] = bson.M{"$regex": tags, "$options": "i"}
		queryParams["tags"] = tags
	}
	if colorTheme := c.QueryParam("color_theme"); colorTheme != "" {
		query["colorTheme"] = colorTheme
		queryParams["color_theme"] = colorTheme
	}
	if ratingMin := c.QueryParam("rating_min"); ratingMin != "" {
		if min, err := strconv.ParseFloat(ratingMin, 64); err == nil {
			query["rating"] = bson.M{"$gte": min}
			queryParams["rating_min"] = ratingMin
		}
	}
	if ratingMax := c.QueryParam("rating_max"); ratingMax != "" {
		if max, err := strconv.ParseFloat(ratingMax, 64); err == nil {
			if existing, ok := query["rating"].(bson.M); ok {
				existing["$lte"] = max
			} else {
				query["rating"] = bson.M{"$lte": max}
			}
			queryParams["rating_max"] = ratingMax
		}
	}
	if isVerified := c.QueryParam("is_verified"); isVerified != "" {
		if isVerified == "true" {
			query["isVerified"] = true
			queryParams["is_verified"] = "true"
		} else if isVerified == "false" {
			query["isVerified"] = false
			queryParams["is_verified"] = "false"
		}
	}
	if listingType := c.QueryParam("listing_type"); listingType != "" {
		query["listingType"] = listingType
		queryParams["listing_type"] = listingType
	}

	page := 1
	limit := 10
	if p := c.QueryParam("page"); p != "" {
		if num, err := strconv.Atoi(p); err == nil && num > 0 {
			page = num
			queryParams["page"] = p
		}
	}
	if l := c.QueryParam("limit"); l != "" {
		if num, err := strconv.Atoi(l); err == nil && num > 0 {
			limit = num
			queryParams["limit"] = l
		}
	}
	skip := (page - 1) * limit

	var properties []models.Property
	cacheKey := utils.GenerateQueryCacheKey("properties", queryParams)
	ctx := context.Background()
	if hit, err := utils.GetCached(ctx, cacheKey, &properties); hit && err == nil {
		return c.JSON(http.StatusOK, properties)
	}

	options := options.Find().SetSkip(int64(skip)).SetLimit(int64(limit))
	cursor, err := pc.collection.Find(ctx, query, options)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch properties"})
	}
	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		var property models.Property
		if err := cursor.Decode(&property); err != nil {
			continue
		}
		properties = append(properties, property)
	}

	if err := utils.SetCached(ctx, cacheKey, properties, 30*time.Second); err != nil {
	}

	return c.JSON(http.StatusOK, properties)
}

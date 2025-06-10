package handlers

import (
	"astrovista-api/cache"
	"astrovista-api/database"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// PostApod fetches the most recent APOD from NASA API and adds it to the database
// @Summary Adds new APOD from NASA
// @Description Fetches the most recent APOD from NASA API and adds it to the database
// @Tags APOD
// @Accept json
// @Produce json
// @Param X-API-Token header string true "Internal API token"
// @Success 201 {object} map[string]interface{}
// @Failure 401 {object} map[string]string
// @Failure 409 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /apod [post]
func PostApod(w http.ResponseWriter, r *http.Request) {
	// Verify basic API token (for internal/scheduled service)
	apiToken := r.Header.Get("X-API-Token")
	if apiToken != os.Getenv("INTERNAL_API_TOKEN") {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Unauthorized - Valid API token required",
		})
		return
	}

	// Get NASA API key from environment variables
	nasaAPIKey := os.Getenv("NASA_API_KEY")
	if nasaAPIKey == "" {
		nasaAPIKey = "DEMO_KEY" // Demo key (limited usage)
	}

	// NASA APOD API URL
	nasaURL := fmt.Sprintf("https://api.nasa.gov/planetary/apod?api_key=%s", nasaAPIKey)

	// Make request to NASA API
	resp, err := http.Get(nasaURL)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error":   "Error fetching data from NASA API",
			"details": err.Error(),
		})
		return
	}
	defer resp.Body.Close()

	// Check if the response was successful
	if resp.StatusCode != http.StatusOK {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error":   "NASA API returned an error",
			"details": resp.Status,
		})
		return
	}

	// Decode JSON response
	var apod Apod
	if err := json.NewDecoder(resp.Body).Decode(&apod); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error":   "Error decoding NASA API response",
			"details": err.Error(),
		})
		return
	}

	// Create context with timeout for MongoDB operations
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Check if a document with this date already exists
	filter := bson.M{"date": apod.Date}
	existingApod := database.ApodCollection.FindOne(ctx, filter)

	// If there was no error, it means a document with this date already exists
	if existingApod.Err() == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict) // 409 Conflict
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error":   "APOD already exists for this date",
			"details": apod.Date,
		})
		return
	} else if existingApod.Err() != mongo.ErrNoDocuments {
		// If the error is different from ErrNoDocuments, there was a database problem
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error":   "Error checking for existing APOD",
			"details": existingApod.Err().Error(),
		})
		return
	}

	// Insert the new document
	result, err := database.ApodCollection.InsertOne(ctx, apod)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error":   "Error inserting APOD into database",
			"details": err.Error(),
		})
		return
	}
	// Invalidate related cache
	// 1. Remove the most recent APOD from cache
	if err := cache.Delete(ctx, "apod:latest"); err != nil {
		log.Printf("Error invalidating cache for the most recent APOD: %v", err)
	}

	// 2. Remove any specific cache for this date
	cacheKey := "apod:date:" + apod.Date
	if err := cache.Delete(ctx, cacheKey); err != nil {
		log.Printf("Error invalidating cache for date %s: %v", apod.Date, err)
	}

	// Return success with the inserted ID
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated) // 201 Created
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "APOD successfully added to database",
		"id":      result.InsertedID,
		"date":    apod.Date,
	})
}

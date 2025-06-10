package handlers

import (
	"astrovista-api/cache"
	"astrovista-api/database"
	"astrovista-api/i18n"
	"astrovista-api/middleware"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// GetApod returns the most recent APOD
// @Summary Get the most recent APOD
// @Description Returns the most recent Astronomy Picture of the Day
// @Tags APOD
// @Accept json
// @Produce json
// @Success 200 {object} Apod
// @Failure 400 {object} map[string]string
// @Router /apod [get]
func GetApod(w http.ResponseWriter, r *http.Request) {
	// Create context with timeout for the database operation
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var apod Apod // struct to store the result

	// Cache key for the most recent APOD
	cacheKey := "apod:latest"

	// Try to retrieve from cache first
	found, err := cache.Get(ctx, cacheKey, &apod)
	if err != nil {
		// Error accessing cache, just log and continue
		log.Printf("Error accessing cache: %v", err)
	}
	// If found in cache, return immediately
	if found {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Cache", "HIT")

		// Get language from request
		lang := middleware.GetLanguageFromContext(r.Context())

		// If not English, try to translate
		if lang != "en" {
			// Convert to map to allow translation
			apodMap := map[string]interface{}{
				"_id":             apod.ID,
				"date":            apod.Date,
				"explanation":     apod.Explanation,
				"hdurl":           apod.Hdurl,
				"media_type":      apod.MediaType,
				"service_version": apod.ServiceVersion,
				"title":           apod.Title,
				"url":             apod.Url,
			}

			// Translate the necessary fields
			if err := i18n.TranslateAPOD(apodMap, lang); err != nil {
				log.Printf("Error translating APOD: %v", err)
			}

			// Send the translated version
			json.NewEncoder(w).Encode(apodMap)
		} else {
			// No translation, send original
			json.NewEncoder(w).Encode(apod)
		}
		return
	}
	// If not found in cache, search in the database
	err = database.ApodCollection.FindOne(
		ctx,
		bson.M{}, // empty filter = all
		options.FindOne().SetSort(bson.D{{Key: "date", Value: -1}}), // sort desc
	).Decode(&apod) // decode the result into the apod variable

	// If found in database, store in cache for future requests
	if err == nil {
		// Store in cache for 1 hour (the most recent may change daily)
		if cacheErr := cache.Set(ctx, cacheKey, apod, 1*time.Hour); cacheErr != nil {
			log.Printf("Error storing in cache: %v", cacheErr)
		}
	}
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Document not found",
		})
		return
	}
	// If found, return JSON to client
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Cache", "MISS") // Indicates it came from the database, not cache

	// Get language from request
	lang := middleware.GetLanguageFromContext(r.Context())

	// If not English, try to translate
	if lang != "en" {
		// Convert to map to allow translation
		apodMap := map[string]interface{}{
			"_id":             apod.ID,
			"date":            apod.Date,
			"explanation":     apod.Explanation,
			"hdurl":           apod.Hdurl,
			"media_type":      apod.MediaType,
			"service_version": apod.ServiceVersion,
			"title":           apod.Title,
			"url":             apod.Url,
		}

		// Translate the necessary fields
		if err := i18n.TranslateAPOD(apodMap, lang); err != nil {
			log.Printf("Error translating APOD: %v", err)
		}

		// Send the translated version
		json.NewEncoder(w).Encode(apodMap)
	} else {
		// No translation, send original
		json.NewEncoder(w).Encode(apod)
	}
}

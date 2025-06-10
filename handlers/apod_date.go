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

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
)

// GetApodDate returns a specific APOD by date
// @Summary Gets an APOD by specific date
// @Description Returns the astronomy picture of the day for the specified date
// @Tags APOD
// @Accept json
// @Produce json
// @Param date path string true "Date in YYYY-MM-DD format" example("2023-01-15")
// @Success 200 {object} Apod
// @Failure 400 {object} map[string]interface{} "Error getting APOD"
// @Router /apod/{date} [get]
func GetApodDate(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	date := params["date"]

	var apod Apod

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Cache key specific for the date
	cacheKey := "apod:date:" + date

	// Try to retrieve from cache first
	found, err := cache.Get(ctx, cacheKey, &apod)
	if err != nil {
		// Error accessing cache, just log and continue
		log.Printf("Error accessing cache: %v", err)
	}
	// If found in cache, apply translation and return
	if found {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Cache", "HIT")

		// Get the language from the request
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

	// Filter: search for document with "date" field equal to the received parameter
	filter := bson.M{"date": date}
	// If not found in cache, search in the database
	err = database.ApodCollection.FindOne(ctx, filter).Decode(&apod)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error":   "Document not found! Please check the date format (YYYY-MM-DD).",
			"details": err.Error(),
		})
		return
	}

	// If found in the database, store in cache for future requests
	// Historical APODs never change, so we can use a long expiration (30 days)
	if cacheErr := cache.Set(ctx, cacheKey, apod, 30*24*time.Hour); cacheErr != nil {
		log.Printf("Error storing in cache: %v", cacheErr)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Cache", "MISS") // Indicates it came from the database, not from cache

	// Get the language from the request
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

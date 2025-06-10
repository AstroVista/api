package handlers

import (
	"astrovista-api/cache"
	"astrovista-api/database"
	"astrovista-api/i18n"
	"astrovista-api/middleware"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

// GetApodsDateRange returns APODs within a date range
// @Summary Get APODs by date range
// @Description Returns the Astronomy Pictures of the Day within a specified date range
// @Tags APODs
// @Accept json
// @Produce json
// @Param start query string false "Start date (YYYY-MM-DD format)" example("2023-01-01")
// @Param end query string false "End date (YYYY-MM-DD format)" example("2023-01-31")
// @Success 200 {object} ApodsDateRangeResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /apods/date-range [get]
func GetApodsDateRange(w http.ResponseWriter, r *http.Request) {
	startDate := r.URL.Query().Get("start")
	endDate := r.URL.Query().Get("end")
	// If "end" parameter is not provided, use the current date
	if endDate == "" {
		endDate = time.Now().Format("2006-01-02")
	}

	// Generate a cache key based on the query parameters
	cacheKey := fmt.Sprintf("apods:range:%s:%s", startDate, endDate)

	// Try to retrieve from cache
	var cachedResponse ApodsDateRangeResponse
	found, err := cache.Get(context.Background(), cacheKey, &cachedResponse)
	if err != nil {
		log.Printf("Error accessing cache for date range: %v", err)
	}
	// If found in cache, return immediately
	if found {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Cache", "HIT")

		// Get language from request
		lang := middleware.GetLanguageFromContext(r.Context())

		// If not English, try to translate each APOD in the result
		if lang != "en" {
			// Create a translated response
			var translatedResponse ApodsDateRangeResponse
			translatedResponse.Count = cachedResponse.Count
			translatedApods := make([]map[string]interface{}, 0, len(cachedResponse.Apods))

			// Translate each APOD
			for _, apod := range cachedResponse.Apods {
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

				translatedApods = append(translatedApods, apodMap)
			}

			// Create a custom response
			customResponse := map[string]interface{}{
				"count": translatedResponse.Count,
				"apods": translatedApods,
			}

			// Send the translated version
			json.NewEncoder(w).Encode(customResponse)
		} else {
			// No translation, send original
			json.NewEncoder(w).Encode(cachedResponse)
		}
		return
	}

	// Verifica se endDate é uma data válida (YYYY-MM-DD)
	if _, err := time.Parse("2006-01-02", endDate); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error":   "Invalid end date format. Use YYYY-MM-DD.",
			"details": err.Error(),
		})
		return
	}

	filter := bson.M{
		"date": bson.M{
			"$gte": startDate,
			"$lte": endDate,
		},
	}
	// If both parameters are empty, return all documents
	if startDate == "" && endDate == "" {
		filter = bson.M{}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := database.ApodCollection.Find(ctx, filter)
	if err != nil {
		fmt.Printf("MongoDB error: %v\n", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error":   "Error fetching documents",
			"details": err.Error(),
		})
		return
	}
	defer cursor.Close(ctx)

	var apods []Apod
	if err = cursor.All(ctx, &apods); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error":   "Error decoding documents",
			"details": err.Error(),
		})
		return
	}

	// Check if no documents were found
	if len(apods) == 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error":   "No documents found for the given date range.",
			"details": fmt.Sprintf("Start date: %s", startDate),
		})
		return
	}

	var response ApodsDateRangeResponse
	response.Count = len(apods)
	if len(apods) == 1 {
		response.Apods = []Apod{apods[0]} // single object as a slice
	} else {
		response.Apods = apods // array of objects
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Size", fmt.Sprintf("%d", len(apods)))
	w.Header().Set("X-Cache", "MISS") // Indicates that it came from the database, not from cache

	// Store in cache for future queries
	// Specific date ranges can be stored for a longer time (12 hours)
	if cacheErr := cache.Set(context.Background(), cacheKey, response, 12*time.Hour); cacheErr != nil {
		log.Printf("Error storing date range in cache: %v", cacheErr)
	}

	// Get language from the request
	lang := middleware.GetLanguageFromContext(r.Context())

	// If not English, try to translate each APOD in the result
	if lang != "en" {
		// Create a translated response
		var translatedResponse ApodsDateRangeResponse
		translatedResponse.Count = response.Count
		translatedApods := make([]map[string]interface{}, 0, len(response.Apods))

		// Translate each APOD
		for _, apod := range response.Apods {
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

			translatedApods = append(translatedApods, apodMap)
		}

		// Create a custom response
		customResponse := map[string]interface{}{
			"count": translatedResponse.Count,
			"apods": translatedApods,
		}

		// Send the translated version
		json.NewEncoder(w).Encode(customResponse)
	} else {
		// No translation, send original
		json.NewEncoder(w).Encode(response)
	}
}

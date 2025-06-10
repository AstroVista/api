package handlers

import (
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

// GetAllApods returns all APODs in the database
// @Summary Get all APODs
// @Description Returns all registered Astronomy Pictures of the Day
// @Tags APODs
// @Accept json
// @Produce json
// @Success 200 {object} AllApodsResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /apods [get]
func GetAllApods(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := database.ApodCollection.Find(ctx, bson.M{})
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
			"error":   "No documents found!",
			"details": "No APODs found in the database.",
		})
		return
	}

	var response AllApodsResponse
	response.Count = len(apods)
	if len(apods) == 1 {
		response.Apods = []Apod{apods[0]} // single object as a slice
	} else {
		response.Apods = apods // array of objects
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Size", fmt.Sprintf("%d", len(apods)))
	// Get language from the request
	lang := middleware.GetLanguageFromContext(r.Context())

	// If not English, try to translate each APOD in the result
	if lang != "en" {
		// Create a translated response
		var translatedResponse AllApodsResponse
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

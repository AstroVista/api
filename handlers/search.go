package handlers

import (
	"astrovista-api/cache"
	"astrovista-api/database"
	"astrovista-api/i18n"
	"astrovista-api/middleware"
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// SearchResponse defined in models.go

// Function to search APODs with various filters and pagination
// @Summary Advanced APOD search
// @Description Search APODs with filters, pagination and sorting
// @Tags APODs
// @Accept json
// @Produce json
// @Param page query int false "Page number" example(1) minimum(1)
// @Param perPage query int false "Items per page (1-200)" example(20) minimum(1) maximum(200)
// @Param mediaType query string false "Media type (image, video or any)" example(image) Enums(image, video, any)
// @Param search query string false "Text to search in title and explanation" example(nebula)
// @Param startDate query string false "Start date (YYYY-MM-DD format)" example(2023-01-01)
// @Param endDate query string false "End date (YYYY-MM-DD format)" example(2023-01-31)
// @Param sort query string false "Sort order (asc or desc)" example(desc) Enums(asc, desc)
// @Success 200 {object} SearchResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /apods/search [get]
func SearchApods(w http.ResponseWriter, r *http.Request) {
	// Create a cache key from the complete query string
	queryHash := md5.Sum([]byte(r.URL.RawQuery))
	cacheKey := "search:" + hex.EncodeToString(queryHash[:])

	// Try to retrieve results from cache
	var cachedResponse SearchResponse
	found, err := cache.Get(r.Context(), cacheKey, &cachedResponse)
	if err != nil {
		log.Printf("Error accessing cache for search: %v", err)
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
			translatedResponse := SearchResponse{
				TotalResults: cachedResponse.TotalResults,
				Page:         cachedResponse.Page,
				PerPage:      cachedResponse.PerPage,
				TotalPages:   cachedResponse.TotalPages,
			}

			translatedApods := make([]map[string]interface{}, 0, len(cachedResponse.Results))

			// Translate each APOD
			for _, apod := range cachedResponse.Results {
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
			} // Create a custom response with standardized fields
			customResponse := map[string]interface{}{
				"total_results": translatedResponse.TotalResults,
				"page":          translatedResponse.Page,
				"per_page":      translatedResponse.PerPage,
				"total_pages":   translatedResponse.TotalPages,
				"results":       translatedApods,
			}

			// Send the translated version
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(customResponse)
		} else {
			// No translation, send original
			json.NewEncoder(w).Encode(cachedResponse)
		}
		return
	}

	// Get parameters from query string
	query := r.URL.Query()
	// Pagination (defaults: page 1, 20 items per page)
	page, err := strconv.Atoi(query.Get("page"))
	if err != nil {
		if query.Get("page") != "" {
			fmt.Printf("Invalid value for page ignored: %s (using 1 as default)\n", query.Get("page"))
		}
		page = 1
	} else if page < 1 {
		fmt.Printf("Invalid value for page (less than 1): %d (using 1 as default)\n", page)
		page = 1
	}
	perPage, err := strconv.Atoi(query.Get("perPage"))
	if err != nil {
		if query.Get("perPage") != "" {
			fmt.Printf("Invalid value for perPage ignored: %s (using 20 as default)\n", query.Get("perPage"))
		}
		perPage = 20 // Default limit
	} else if perPage < 1 || perPage > 200 {
		fmt.Printf("Value out of bounds for perPage: %d (must be between 1 and 200, using 20 as default)\n", perPage)
		perPage = 20 // Default limit
	}
	// Building the MongoDB filter
	filter := bson.M{}
	// Various filters - validates that mediaType is only "image" or "video"
	if mediaType := query.Get("mediaType"); mediaType != "" && mediaType != "any" {
		// Checks if the value belongs to the allowed enum
		if mediaType == "image" || mediaType == "video" {
			filter["media_type"] = mediaType
		} else {
			// If invalid value, ignore the filter (as if it was not provided)
			fmt.Printf("Invalid value for mediaType ignored: %s\n", mediaType)
		}
	}

	// Text search (in title and explanation)
	if search := query.Get("search"); search != "" {
		// Text search in multiple fields
		textFilter := bson.M{
			"$or": []bson.M{
				{"title": bson.M{"$regex": search, "$options": "i"}},
				{"explanation": bson.M{"$regex": search, "$options": "i"}},
			},
		}

		// If there are already other filters, combine with them
		if len(filter) > 0 {
			filter = bson.M{
				"$and": []bson.M{
					filter,
					textFilter,
				},
			}
		} else {
			filter = textFilter
		}
	}
	// Date filter
	if startDate := query.Get("startDate"); startDate != "" {
		if _, err := time.Parse("2006-01-02", startDate); err == nil {
			if endDate := query.Get("endDate"); endDate != "" {
				if _, err := time.Parse("2006-01-02", endDate); err == nil {
					filter["date"] = bson.M{
						"$gte": startDate,
						"$lte": endDate,
					}
				} else {
					// Invalid date format for endDate
					fmt.Printf("Invalid date format for endDate ignored: %s\n", endDate)
				}
			} else {
				filter["date"] = bson.M{"$gte": startDate}
			}
		} else {
			// Invalid date format for startDate
			fmt.Printf("Invalid date format for startDate ignored: %s\n", startDate)
		}
	} else if endDate := query.Get("endDate"); endDate != "" {
		if _, err := time.Parse("2006-01-02", endDate); err == nil {
			filter["date"] = bson.M{"$lte": endDate}
		} else {
			// Invalid date format for endDate
			fmt.Printf("Invalid date format for endDate ignored: %s\n", endDate)
		}
	}
	// Sorting (default: descending date / most recent first)
	sortDirection := -1 // -1 = desc, 1 = asc
	if sort := query.Get("sort"); sort != "" {
		// Converts to lowercase for case-insensitive comparison
		sortLower := strings.ToLower(sort)
		// Validates if it's an allowed value
		if sortLower == "asc" {
			sortDirection = 1
		} else if sortLower == "desc" {
			sortDirection = -1 // keeps the default
		} else {
			// Ignores invalid value and keeps the default
			fmt.Printf("Invalid value for sort ignored: %s (using 'desc' as default)\n", sort)
		}
	}

	// Setting up sorting and pagination
	findOptions := options.Find().
		SetSort(bson.D{{Key: "date", Value: sortDirection}}).
		SetSkip(int64((page - 1) * perPage)).
		SetLimit(int64(perPage))

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// First, count the total number of documents to calculate pagination
	totalResults, err := database.ApodCollection.CountDocuments(ctx, filter)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error":   "Error counting documents",
			"details": err.Error(),
		})
		return
	}

	// Next, fetch the documents of the current page
	cursor, err := database.ApodCollection.Find(ctx, filter, findOptions)
	if err != nil {
		fmt.Printf("MongoDB search error: %v\n", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error":   "Error searching documents",
			"details": err.Error(),
		})
		return
	}
	defer cursor.Close(ctx)

	// Decodes the results
	var apods []Apod
	if err = cursor.All(ctx, &apods); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error":   "Error decoding search results",
			"details": err.Error(),
		})
		return
	}
	// Checks if any results were found
	if len(apods) == 0 && page == 1 {
		// Debug log showing the filters used
		filterJSON, _ := json.Marshal(filter)
		fmt.Printf("No results found for filter: %s\n", string(filterJSON))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "No documents found matching the search criteria",
		})
		return
	}

	// Calculates the total number of pages
	totalPages := int(math.Ceil(float64(totalResults) / float64(perPage)))

	// Prepares the response
	response := SearchResponse{
		TotalResults: int(totalResults),
		Page:         page,
		PerPage:      perPage,
		TotalPages:   totalPages,
		Results:      apods,
	} // Stores the response in the cache with an expiration of 5 minutes
	if err := cache.Set(r.Context(), cacheKey, response, 5*time.Minute); err != nil {
		log.Printf("Error storing in cache: %v", err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Cache", "MISS")

	// Get language from request
	lang := middleware.GetLanguageFromContext(r.Context())

	// If not English, try to translate each APOD in the result
	if lang != "en" {
		translatedApods := make([]map[string]interface{}, 0, len(response.Results))

		// Translate each APOD
		for _, apod := range response.Results {
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
			"totalResults": response.TotalResults,
			"page":         response.Page,
			"perPage":      response.PerPage,
			"totalPages":   response.TotalPages,
			"results":      translatedApods,
		}

		// Send the translated version
		json.NewEncoder(w).Encode(customResponse)
	} else {
		// No translation, send original
		json.NewEncoder(w).Encode(response)
	}
}

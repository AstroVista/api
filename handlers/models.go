package handlers

import (
	"encoding/json"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Apod represents an APOD record from NASA
// swagger:model Apod
type Apod struct {
	// MongoDB ID
	// example: 507f1f77bcf86cd799439011
	ID primitive.ObjectID `bson:"_id,omitempty" json:"_id"`
	// Date in string format (e.g. "1995-06-16")
	// example: 2023-01-15
	// format: date
	Date string `bson:"date" json:"date"`
	// Explanation of the astronomy picture of the day
	// example: A beautiful nebula captured by the Hubble telescope
	Explanation string `bson:"explanation" json:"explanation"`
	// URL of the high-definition image
	// example: https://apod.nasa.gov/apod/image/2301/M31_HubbleSpitzerGendler_960.jpg
	// format: uri
	Hdurl string `bson:"hdurl" json:"hdurl"`
	// Media type (image or video)
	// example: image
	// enum: image,video
	MediaType string `bson:"media_type" json:"media_type"`
	// API service version
	// example: v1
	ServiceVersion string `bson:"service_version" json:"service_version"`
	// Title of the astronomy picture of the day
	// example: Andromeda Galaxy
	Title string `bson:"title" json:"title"`
	// URL of the standard resolution image
	// example: https://apod.nasa.gov/apod/image/2301/M31_HubbleSpitzerGendler_960.jpg
	// format: uri
	Url string `bson:"url" json:"url"`
}

// AllApodsResponse is the response structure for endpoints that return multiple APODs
// swagger:model AllApodsResponse
type AllApodsResponse struct {
	// Total number of APODs found
	// example: 15
	Count int `json:"count"`
	// List of APODs
	Apods []Apod `json:"apods"`
}

// ApodsDateRangeResponse is the response structure for date range search
// swagger:model ApodsDateRangeResponse
type ApodsDateRangeResponse struct {
	// Total number of APODs found
	// example: 7
	Count int `json:"count"`
	// List of APODs
	Apods []Apod `json:"apods"`
}

// SearchResponse is the response structure for the search endpoint
// swagger:model SearchResponse
type SearchResponse struct {
	// Total number of results found
	// example: 42
	TotalResults int `json:"total_results"` // Using snake_case for consistency
	// Current page number
	// example: 1
	Page int `json:"page"`
	// Items per page
	// example: 20
	PerPage int `json:"per_page"` // Using snake_case for consistency
	// Total number of available pages
	// example: 3
	TotalPages int `json:"total_pages"` // Using snake_case for consistency
	// Search results
	Results []Apod `json:"results"`
}

// MarshalJSON customizes JSON serialization to support translation
func (a Apod) MarshalJSON() ([]byte, error) {
	// Create a map with the APOD fields
	apodMap := map[string]interface{}{
		"_id":             a.ID,
		"date":            a.Date,
		"explanation":     a.Explanation,
		"hdurl":           a.Hdurl,
		"media_type":      a.MediaType,
		"service_version": a.ServiceVersion,
		"title":           a.Title,
		"url":             a.Url,
	}

	// In standard serialization we don't do anything
	// Translation will be applied in handlers before calling json.Marshal

	return json.Marshal(apodMap)
}

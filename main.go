package main

import (
	"astrovista-api/cache"
	"astrovista-api/database"
	_ "astrovista-api/docs" // Importing docs for Swagger
	"astrovista-api/handlers"
	"astrovista-api/i18n"
	"astrovista-api/middleware"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	httpSwagger "github.com/swaggo/http-swagger"
)

// @title           AstroVista API
// @version         1.0
// @description     API for managing NASA APOD (Astronomy Picture of the Day) data
// @BasePath        /
func main() {
	// Initialize database and cache connections
	database.Connect()
	cache.Connect()
	// Initialize internationalization system
	i18n.InitLocales()
	i18n.InitTranslationService()

	router := mux.NewRouter()
	// Swagger configuration
	router.PathPrefix("/swagger/").Handler(httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"), // URL to access JSON documentation
		httpSwagger.DeepLinking(true),
		httpSwagger.DocExpansion("list"),
		httpSwagger.DomID("swagger-ui"),
	))

	// Redirect root route to Swagger documentation
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/swagger/", http.StatusFound)
	})

	// Public GET endpoints (no rate limit) // Add middleware for JSON formatting and language detection
	router.Use(middleware.JSONFormatterMiddleware)
	router.Use(middleware.LanguageDetector)
	router.HandleFunc("/apod", handlers.GetApod).Methods("GET")
	router.HandleFunc("/apod/{date}", handlers.GetApodDate).Methods("GET")
	router.HandleFunc("/apods", handlers.GetAllApods).Methods("GET")
	router.HandleFunc("/apods/search", handlers.SearchApods).Methods("GET")
	router.HandleFunc("/apods/date-range", handlers.GetApodsDateRange).Methods("GET")
	router.HandleFunc("/languages", handlers.GetSupportedLanguages).Methods("GET")
	// Rate limiter: 1 request per minute
	rateLimiter := middleware.NewRateLimiter(1, 1*time.Minute)

	// POST endpoint with applied rate limit
	postRouter := router.PathPrefix("/apod").Subrouter()
	postRouter.Use(rateLimiter.Limit)
	postRouter.HandleFunc("", handlers.PostApod).Methods("POST")
	// Determine server port (default 8080, or use PORT environment variable)
	port := "8080"

	log.Printf("Server running on port %s!", port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}

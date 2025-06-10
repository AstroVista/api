package i18n

import (
	"astrovista-api/cache"
	"fmt"
	"log"
	"strings"
)

// TranslationService defines the interface for translation services
type TranslationService interface {
	Translate(text, sourceLang, targetLang string) (string, error)
}

// mockTranslationService is a mock implementation for development
type mockTranslationService struct{}

// Translate in the mock implementation just adds a language indicator
func (s *mockTranslationService) Translate(text, sourceLang, targetLang string) (string, error) {
	if len(text) > 100 {
		// For long explanations, we truncate for simulation
		return fmt.Sprintf("%s... [Translated to %s]", text[:100], targetLang), nil
	}
	return fmt.Sprintf("%s [%s]", text, targetLang), nil
}

// googleTranslationService would be a real implementation using the Google Translate API
type googleTranslationService struct {
	apiKey string
}

// Translate in the Google implementation (sketch)
func (s *googleTranslationService) Translate(text, sourceLang, targetLang string) (string, error) {
	// Here would be the code to call the Google Translate API
	// For now, we just simulate
	log.Printf("Simulating translation of '%s' from '%s' to '%s'",
		truncateForLogging(text), sourceLang, targetLang)
	return text + " [Google Translated]", nil
}

// deepLTranslationService would be a real implementation using the DeepL API
type deepLTranslationService struct {
	apiKey string
}

// Translate in the DeepL implementation (sketch)
func (s *deepLTranslationService) Translate(text, sourceLang, targetLang string) (string, error) {
	// Here would be the code to call the DeepL API
	// For now, we just simulate
	log.Printf("Simulating DeepL translation of '%s' from '%s' to '%s'",
		truncateForLogging(text), sourceLang, targetLang)
	return text + " [DeepL Translated]", nil
}

// Current translation service
var currentService TranslationService

// InitTranslationService initializes the appropriate translation service
func InitTranslationService() {
	// Check which service to use based on environment variables
	if apiKey := GoogleTranslateAPIKey(); apiKey != "" {
		log.Println("Using Google Translate for translations")
		googleClient := NewGoogleTranslateClient(apiKey)

		// Enable Redis cache if available
		if cache.Client != nil {
			log.Println("Redis cache enabled for translations")
			googleClient.cache.EnableRedisCache()
		}

		currentService = googleClient
	} else if apiKey := DeepLAPIKey(); apiKey != "" {
		log.Println("Using DeepL for translations")
		deepLClient := NewDeepLClient(apiKey)

		// Enable Redis cache if available
		if cache.Client != nil {
			log.Println("Redis cache enabled for translations")
			deepLClient.cache.EnableRedisCache()
		}

		currentService = deepLClient
	} else {
		log.Println("No translation API configured, using mock service")
		currentService = &mockTranslationService{}
	}
}

// TranslateText translates the text to the target language
func TranslateText(text, targetLang string) (string, error) {
	if currentService == nil {
		InitTranslationService()
	}

	// If the target language is English or empty, we do not translate
	if targetLang == "" || targetLang == "en" {
		return text, nil
	}

	// Truncate very long text (just for logging, not for actual translation)
	logText := truncateForLogging(text)
	log.Printf("Translating text: '%s' to '%s'", logText, targetLang)

	// Assume English as the source language
	return currentService.Translate(text, "en", targetLang)
}

// Helper method to truncate long text in logs
func truncateForLogging(text string) string {
	if len(text) > 50 {
		return text[:50] + "..."
	}
	return text
}

// TryTranslate tries to translate a text, returning the original in case of error
func TryTranslate(text string, targetLang string) string {
	if targetLang == "en" || strings.TrimSpace(text) == "" {
		return text
	}

	translated, err := TranslateText(text, targetLang)
	if err != nil {
		log.Printf("Error translating text: %v", err)
		return text
	}
	return translated
}

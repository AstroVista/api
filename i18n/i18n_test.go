package i18n

import (
	"astrovista-api/i18n"
	"os"
	"testing"
)

// TestTranslationService tests the different translation services
func TestTranslationService(t *testing.T) {
	// Backup of original environment variables
	originalGoogleKey := os.Getenv("GOOGLE_TRANSLATE_API_KEY")
	originalDeepLKey := os.Getenv("DEEPL_API_KEY")

	// Clear the variables at the end of the test
	defer func() {
		os.Setenv("GOOGLE_TRANSLATE_API_KEY", originalGoogleKey)
		os.Setenv("DEEPL_API_KEY", originalDeepLKey)
	}()

	// Test the Mock service (default when no keys are configured)
	t.Run("MockTranslationService", func(t *testing.T) {
		// Clear variables to ensure mock will be used
		os.Setenv("GOOGLE_TRANSLATE_API_KEY", "")
		os.Setenv("DEEPL_API_KEY", "")

		// Reinitialize the service
		i18n.InitTranslationService()

		text := "Hello, world!"
		translated, err := i18n.TranslateText(text, "pt-BR")

		if err != nil {
			t.Errorf("Error translating text: %v", err)
		}

		// Check if the text was modified (mock adds [pt-BR] at the end)
		if translated == text {
			t.Errorf("Text was not translated by mock")
		}
	})

	// Tests integrating with real APIs could be added here,
	// but would not be executed in automated CI (would need real keys)
}

// TestTranslateAPOD tests the translation of fields in an APOD document
func TestTranslateAPOD(t *testing.T) {
	// Create a fictional APOD for testing
	apodData := map[string]interface{}{
		"title":       "Amazing Galaxy",
		"explanation": "This is a beautiful galaxy far away.",
	}

	// Restart the translation service to use the mock
	os.Setenv("GOOGLE_TRANSLATE_API_KEY", "")
	os.Setenv("DEEPL_API_KEY", "")
	i18n.InitTranslationService()

	// Test with English (should not modify)
	i18n.TranslateAPOD(apodData, "en")
	if apodData["title"] != "Amazing Galaxy" {
		t.Errorf("The text in English should not be modified")
	}

	// Test with Portuguese
	i18n.TranslateAPOD(apodData, "pt-BR")

	// With the mock, it should have added [pt-BR] to the title
	if title, ok := apodData["title"].(string); !ok || title == "Amazing Galaxy" {
		t.Errorf("The text should have been translated, but it remains: %v", apodData["title"])
	}
}

// TestTranslationCache tests the functioning of the translation cache
func TestTranslationCache(t *testing.T) {
	cache := i18n.NewTranslationCache()

	// Test add and retrieve from cache
	testKey := "test:en:pt-BR:hello"
	testValue := "olá"

	// Initially, the value should not exist
	_, found := cache.Get(testKey)
	if found {
		t.Errorf("Value should not exist in cache yet")
	}

	// Add to cache
	cache.Set(testKey, testValue)

	// Now it should be found
	value, found := cache.Get(testKey)
	if !found {
		t.Errorf("Value should exist in cache after Set")
	}

	if value != testValue {
		t.Errorf("Retrieved value (%s) does not match stored value (%s)", value, testValue)
	}

	// Test clearing
	cache.Clear()
	_, found = cache.Get(testKey)
	if found {
		t.Errorf("Value still exists in cache after Clear")
	}
}

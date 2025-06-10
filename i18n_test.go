package main

import (
	"astrovista-api/i18n"
	"astrovista-api/middleware"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestTranslationMiddleware verifies if the language detection middleware works correctly
func TestTranslationMiddleware(t *testing.T) {
	// Initialize i18n system
	i18n.InitLocales()

	// Create a test handler that simply returns the detected language
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lang := middleware.GetLanguageFromContext(r.Context())
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"detectedLanguage": lang})
	})

	// Apply the middleware
	handlerWithMiddleware := middleware.LanguageDetector(testHandler)

	// Tests for different language detection methods
	testCases := []struct {
		name           string
		acceptLanguage string
		queryParam     string
		expected       string
	}{
		{
			name:           "Default language",
			acceptLanguage: "",
			queryParam:     "",
			expected:       "en",
		},
		{
			name:           "Accept-Language header",
			acceptLanguage: "pt-BR",
			queryParam:     "",
			expected:       "pt-BR",
		},
		{
			name:           "Query parameter",
			acceptLanguage: "",
			queryParam:     "es",
			expected:       "es",
		},
		{
			name:           "Query parameter overrides header",
			acceptLanguage: "pt-BR",
			queryParam:     "fr",
			expected:       "fr",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Cria uma requisição de teste
			req, err := http.NewRequest("GET", "/", nil)
			if err != nil {
				t.Fatal(err)
			}

			// Configura os parâmetros de teste
			if tc.acceptLanguage != "" {
				req.Header.Set("Accept-Language", tc.acceptLanguage)
			}
			if tc.queryParam != "" {
				q := req.URL.Query()
				q.Add("lang", tc.queryParam)
				req.URL.RawQuery = q.Encode()
			}

			// Executa a requisição
			rr := httptest.NewRecorder()
			handlerWithMiddleware.ServeHTTP(rr, req)

			// Verifica o status code
			if rr.Code != http.StatusOK {
				t.Errorf("Status code esperado %d, obtido %d", http.StatusOK, rr.Code)
			}

			// Verifica o idioma retornado
			var response map[string]string
			if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
				t.Fatal(err)
			}

			if response["detectedLanguage"] != tc.expected {
				t.Errorf("Expected language %q, got %q", tc.expected, response["detectedLanguage"])
			}
		})
	}
}

// TestTranslateAPOD tests the translation of APOD data
func TestTranslateAPOD(t *testing.T) {
	// Initialize i18n system
	i18n.InitLocales()
	i18n.InitTranslationService()

	// Create a test APOD
	apodData := map[string]interface{}{
		"title":       "The Milky Way over the Grand Canyon",
		"explanation": "This is a stunning view of our galaxy spanning over the Grand Canyon.",
		"date":        "2023-01-15",
		"media_type":  "image",
		"url":         "https://example.com/image.jpg",
	}

	// Test translation for different languages
	languages := []string{"pt-BR", "es", "fr"}

	for _, lang := range languages {
		t.Run("Translation to "+lang, func(t *testing.T) {
			// Create a copy of the data to not affect subsequent tests
			apodCopy := make(map[string]interface{})
			for k, v := range apodData {
				apodCopy[k] = v
			}

			// Apply the translation
			err := i18n.TranslateAPOD(apodCopy, lang)
			if err != nil {
				t.Fatalf("Error translating APOD: %v", err)
			}

			// Check if fields were translated
			origTitle := apodData["title"].(string)
			transTitle := apodCopy["title"].(string)

			if transTitle == origTitle {
				t.Logf("Warning: title was not changed. This is expected if no translation API is configured.")
			} else {
				t.Logf("Original title: %q", origTitle)
				t.Logf("Translated title: %q", transTitle)
			}

			origExplanation := apodData["explanation"].(string)
			transExplanation := apodCopy["explanation"].(string)

			if transExplanation == origExplanation {
				t.Logf("Warning: explanation was not changed. This is expected if no translation API is configured.")
			} else {
				t.Logf("First 50 characters of original explanation: %q", origExplanation[:min(50, len(origExplanation))])
				t.Logf("First 50 characters of translated explanation: %q", transExplanation[:min(50, len(transExplanation))])
			}
		})
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

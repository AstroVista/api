package i18n

import (
	"astrovista-api/cache"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

// GoogleTranslateRequest represents the request format for the API
type GoogleTranslateRequest struct {
	Q      []string `json:"q"`
	Source string   `json:"source"`
	Target string   `json:"target"`
	Format string   `json:"format"`
}

// GoogleTranslateResponse represents the API response
type GoogleTranslateResponse struct {
	Data struct {
		Translations []struct {
			TranslatedText string `json:"translatedText"`
		} `json:"translations"`
	} `json:"data"`
}

// GoogleTranslateClient implements the translation service using Google Translate API
type GoogleTranslateClient struct {
	apiKey     string
	httpClient *http.Client
	cache      *TranslationCache
}

// NewGoogleTranslateClient creates a new client for the Google Translate API
func NewGoogleTranslateClient(apiKey string) *GoogleTranslateClient { // Use Redis cache if available, otherwise use in-memory cache
	var translationCache *TranslationCache
	if cache.Client != nil {
		log.Println("Google Translate using Redis cache for translations")
		translationCache = NewTranslationCache()
		translationCache.EnableRedisCache()
	} else {
		log.Println("Google Translate using in-memory cache for translations")
		translationCache = NewTranslationCache()
	}

	return &GoogleTranslateClient{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		cache: translationCache,
	}
}

// Translate implements the TranslationService interface for Google Translate
func (c *GoogleTranslateClient) Translate(text, sourceLang, targetLang string) (string, error) {
	// Check the cache first
	cacheKey := fmt.Sprintf("%s:%s:%s", sourceLang, targetLang, getHashKey(text))
	if cachedText, found := c.cache.Get(cacheKey); found {
		return cachedText, nil
	}

	// Sanitize languages to the format expected by Google
	sourceLang = sanitizeLanguageCode(sourceLang)
	targetLang = sanitizeLanguageCode(targetLang)

	// Prepare the request
	reqBody := GoogleTranslateRequest{
		Q:      []string{text},
		Source: sourceLang,
		Target: targetLang,
		Format: "text", // or "html" if the text contains HTML
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("error serializing request: %v", err)
	}

	// API URL with API key
	url := fmt.Sprintf("https://translation.googleapis.com/language/translate/v2?key=%s", c.apiKey)

	// Create the HTTP request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("error creating request: %v", err)
	}

	// Set the headers
	req.Header.Set("Content-Type", "application/json")
	// Execute the request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("error executing request: %v", err)
	}
	defer resp.Body.Close()

	// Check the response status
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API returned non-OK status: %d", resp.StatusCode)
	}

	// Decode the response
	var translateResp GoogleTranslateResponse
	if err := json.NewDecoder(resp.Body).Decode(&translateResp); err != nil {
		return "", fmt.Errorf("error decoding response: %v", err)
	}

	// Check if there are translations
	if len(translateResp.Data.Translations) == 0 {
		return "", fmt.Errorf("no translation returned")
	}

	// Get the translated text
	translatedText := translateResp.Data.Translations[0].TranslatedText

	// Store in cache
	c.cache.Set(cacheKey, translatedText)

	return translatedText, nil
}

// Sanitize the language code to the format accepted by Google Translate
func sanitizeLanguageCode(lang string) string {
	// Google Translate uses simple codes like "pt" instead of "pt-BR"
	parts := strings.Split(lang, "-")
	return strings.ToLower(parts[0])
}

// getHashKey creates a simplified hash key for long texts
func getHashKey(text string) string {
	if len(text) <= 32 {
		return text
	}
	// A simple implementation for long texts
	return fmt.Sprintf("%s...%s:%d", text[:16], text[len(text)-16:], len(text))
}

// GoogleTranslateAPIKey returns the Google Translate API key from the environment
func GoogleTranslateAPIKey() string {
	return os.Getenv("GOOGLE_TRANSLATE_API_KEY")
}

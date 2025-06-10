package i18n

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

// DeepLTranslateRequest represents the request format for the DeepL API
type DeepLTranslateRequest struct {
	Text       []string `json:"text"`
	SourceLang string   `json:"source_lang,omitempty"`
	TargetLang string   `json:"target_lang"`
	Formality  string   `json:"formality,omitempty"`
}

// DeepLTranslateResponse represents the response from the DeepL API
type DeepLTranslateResponse struct {
	Translations []struct {
		Text string `json:"text"`
	} `json:"translations"`
}

// DeepLClient implements the translation service using the DeepL API
type DeepLClient struct {
	apiKey     string
	httpClient *http.Client
	cache      *TranslationCache
	freeAPI    bool // Indicates whether using the free API or Pro
}

// NewDeepLClient creates a new client for the DeepL API
func NewDeepLClient(apiKey string) *DeepLClient {
	// DeepL distinguishes between free and Pro API by the key prefix
	isFreeAPI := strings.HasPrefix(apiKey, "DeepL-Auth-Key ")

	return &DeepLClient{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		cache:   NewTranslationCache(),
		freeAPI: isFreeAPI,
	}
}

// Translate implements the TranslationService interface for DeepL
func (c *DeepLClient) Translate(text, sourceLang, targetLang string) (string, error) {
	// Check the cache first
	cacheKey := fmt.Sprintf("deepl:%s:%s:%s", sourceLang, targetLang, getHashKey(text))
	if cachedText, found := c.cache.Get(cacheKey); found {
		return cachedText, nil
	}

	// Adapt the language code to the format expected by DeepL
	targetLang = adaptLanguageForDeepL(targetLang)

	// Prepare the request
	reqBody := DeepLTranslateRequest{
		Text:       []string{text},
		TargetLang: targetLang,
	}

	// Only define the source language if specified
	if sourceLang != "" {
		reqBody.SourceLang = adaptLanguageForDeepL(sourceLang)
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("error serializing request: %v", err)
	}

	// API URL depending on the type (Free or Pro)
	var url string
	if c.freeAPI {
		url = "https://api-free.deepl.com/v2/translate"
	} else {
		url = "https://api.deepl.com/v2/translate"
	}
	// Create the HTTP request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("error creating request: %v", err)
	}

	// Set the headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", c.apiKey)

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
	var translateResp DeepLTranslateResponse
	if err := json.NewDecoder(resp.Body).Decode(&translateResp); err != nil {
		return "", fmt.Errorf("error decoding response: %v", err)
	}

	// Check if there are translations
	if len(translateResp.Translations) == 0 {
		return "", fmt.Errorf("no translation returned")
	}
	// Get the translated text
	translatedText := translateResp.Translations[0].Text

	// Store in cache
	c.cache.Set(cacheKey, translatedText)

	return translatedText, nil
}

// adaptLanguageForDeepL converts language codes to the format expected by DeepL
func adaptLanguageForDeepL(lang string) string {
	// DeepL uses codes like "PT-BR", "EN-US" (uppercase)
	parts := strings.Split(lang, "-")
	if len(parts) == 1 {
		return strings.ToUpper(parts[0])
	}

	// For compound codes, capitalize both parts
	return strings.ToUpper(parts[0]) + "-" + strings.ToUpper(parts[1])
}

// DeepLAPIKey returns the DeepL API key from the environment
func DeepLAPIKey() string {
	return os.Getenv("DEEPL_API_KEY")
}

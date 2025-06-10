package i18n

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

var (
	// Bundle contains all translated messages
	Bundle *i18n.Bundle
	// Languages supported by the API
	SupportedLanguages = []string{"en", "pt-BR", "es", "fr", "de", "it"}
)

// InitLocales initializes the internationalization system
func InitLocales() {
	// Create a new bundle with English as the base language
	Bundle = i18n.NewBundle(language.English)
	Bundle.RegisterUnmarshalFunc("json", json.Unmarshal)

	// Load translation files
	if err := loadTranslationFiles(); err != nil {
		log.Printf("Warning: Failed to load translations: %v", err)
		log.Println("API will only work with English language")
	}
}

// LoadTranslationFiles loads translation files from the i18n/locales directory
func loadTranslationFiles() error {
	localesDir := filepath.Join("i18n", "locales")
	// Check if directory exists
	if _, err := os.ReadDir(localesDir); err != nil {
		// If directory doesn't exist, create it and add default translation files
		if err := createDefaultTranslationFiles(localesDir); err != nil {
			return err
		}
	}

	// Load all translation files
	for _, lang := range SupportedLanguages {
		filename := filepath.Join(localesDir, lang+".json")
		if _, err := Bundle.LoadMessageFile(filename); err != nil {
			log.Printf("Error loading translation for %s: %v", lang, err)
		}
	}

	return nil
}

// createDefaultTranslationFiles creates default translation files if they don't exist
func createDefaultTranslationFiles(localesDir string) error {
	// Create the directory if it doesn't exist
	if err := os.MkdirAll(localesDir, 0755); err != nil {
		return err
	}

	// Create translation files with basic content
	translations := map[string]map[string]string{
		"en": {
			"apod_title":        "Astronomy Picture of the Day",
			"apod_not_found":    "Document not found! Please check the date format (YYYY-MM-DD).",
			"search_no_results": "No documents found matching the search criteria",
		},
		"pt-BR": {
			"apod_title":        "Imagem Astronômica do Dia",
			"apod_not_found":    "Documento não encontrado! Por favor verifique o formato da data (AAAA-MM-DD).",
			"search_no_results": "Nenhum documento encontrado para os critérios de busca",
		},
		"es": {
			"apod_title":        "Imagen Astronómica del Día",
			"apod_not_found":    "¡Documento no encontrado! Por favor verifique el formato de la fecha (AAAA-MM-DD).",
			"search_no_results": "No se encontraron documentos que coincidan con los criterios de búsqueda",
		},
		"fr": {
			"apod_title":        "Image Astronomique du Jour",
			"apod_not_found":    "Document non trouvé! Veuillez vérifier le format de la date (AAAA-MM-JJ).",
			"search_no_results": "Aucun document trouvé correspondant aux critères de recherche",
		},
	}

	for lang, msgs := range translations {
		filename := filepath.Join(localesDir, lang+".json")

		// Convert to the format expected by go-i18n
		i18nMsgs := make(map[string]map[string]string)
		for id, msg := range msgs {
			i18nMsgs[id] = map[string]string{"other": msg}
		}

		// Serialize to JSON
		data, err := json.MarshalIndent(i18nMsgs, "", "  ")
		if err != nil {
			return err
		}
		// Write to file
		if err := os.WriteFile(filename, data, 0644); err != nil {
			return err
		}
	}

	return nil
}

// Localizer returns a localizer for the specified language
func Localizer(lang string) *i18n.Localizer {
	// If language is not specified or not supported, use English
	if lang == "" {
		lang = "en"
	}

	// Check if language is supported
	supported := false
	for _, supportedLang := range SupportedLanguages {
		if strings.HasPrefix(lang, supportedLang) {
			supported = true
			break
		}
	}

	if !supported {
		lang = "en"
	}

	return i18n.NewLocalizer(Bundle, lang, "en")
}

// TranslateAPOD translates APOD fields to the requested language
func TranslateAPOD(apodData map[string]interface{}, lang string) error {
	// If not a supported language or it's English, return without modifications
	if lang == "" || lang == "en" {
		return nil
	}

	// Initialize translation service if not yet initialized
	if currentService == nil {
		InitTranslationService()
	}

	// Translate title
	if title, ok := apodData["title"].(string); ok && title != "" {
		translatedTitle, err := TranslateText(title, lang)
		if err == nil {
			apodData["title"] = translatedTitle
		} else {
			log.Printf("Error translating title: %v", err)
		}
	}

	// Translate explanation
	if explanation, ok := apodData["explanation"].(string); ok && explanation != "" {
		translatedExplanation, err := TranslateText(explanation, lang)
		if err == nil {
			apodData["explanation"] = translatedExplanation
		} else {
			log.Printf("Error translating explanation: %v", err)
		}
	}

	// Other fields can be translated here if necessary
	// For example, copyright, etc.
	if copyright, ok := apodData["copyright"].(string); ok && copyright != "" {
		translatedCopyright, err := TranslateText(copyright, lang)
		if err == nil {
			apodData["copyright"] = translatedCopyright
		}
	}

	return nil
}

// Helper function to get the minimum between two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

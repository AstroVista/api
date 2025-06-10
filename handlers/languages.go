package handlers

import (
	"astrovista-api/i18n"
	"encoding/json"
	"net/http"
)

// LanguageInfo contains information about a supported language
type LanguageInfo struct {
	Code       string `json:"code"`
	Name       string `json:"name"`
	NativeName string `json:"nativeName"`
}

// GetSupportedLanguages returns the list of languages supported by the API
// @Summary List supported languages
// @Description Returns the list of languages supported by the AstroVista API
// @Tags Configuration
// @Accept json
// @Produce json
// @Success 200 {array} LanguageInfo
// @Router /languages [get]
func GetSupportedLanguages(w http.ResponseWriter, r *http.Request) { // Maps language codes to their names
	languageNames := map[string]map[string]string{
		"en": {
			"name":       "English",
			"nativeName": "English",
		},
		"pt-BR": {
			"name":       "Brazilian Portuguese",
			"nativeName": "Português do Brasil",
		},
		"es": {
			"name":       "Spanish",
			"nativeName": "Español",
		},
		"fr": {
			"name":       "French",
			"nativeName": "Français",
		},
		"de": {
			"name":       "German",
			"nativeName": "Deutsch",
		},
		"it": {
			"name":       "Italian",
			"nativeName": "Italiano",
		},
	}

	// Prepare the list of supported languages
	var languages []LanguageInfo

	for _, lang := range i18n.SupportedLanguages {
		info, exists := languageNames[lang]
		if !exists {
			info = map[string]string{
				"name":       lang,
				"nativeName": lang,
			}
		}

		languages = append(languages, LanguageInfo{
			Code:       lang,
			Name:       info["name"],
			NativeName: info["nativeName"],
		})
	}

	// Return the list in JSON format
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(languages)
}

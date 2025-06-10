package middleware

import (
	"context"
	"net/http"
	"strings"
)

// Context key to store the language
type langKey struct{}

// LanguageDetector is a middleware that detects the user's preferred language
func LanguageDetector(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get language from Accept-Language header
		acceptLang := r.Header.Get("Accept-Language")

		// Get language from 'lang' query string (takes precedence over header)
		queryLang := r.URL.Query().Get("lang")
		if queryLang != "" {
			acceptLang = queryLang
		}

		// Extract the main language code (e.g., "pt-BR" -> "pt")
		lang := "en" // default
		if acceptLang != "" {
			parts := strings.Split(acceptLang, ",")
			langParts := strings.Split(parts[0], ";") // Remove q-factor
			lang = strings.TrimSpace(langParts[0])
		}

		// Store the language in the request context
		ctx := context.WithValue(r.Context(), langKey{}, lang)

		// Call the next handler with the updated context
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetLanguageFromContext extracts the language from the request context
func GetLanguageFromContext(ctx context.Context) string {
	lang, ok := ctx.Value(langKey{}).(string)
	if !ok {
		return "en" // default language
	}
	return lang
}

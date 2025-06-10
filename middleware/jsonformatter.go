package middleware

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strconv"
)

// JSONResponseWriter is a wrapper for http.ResponseWriter that formats JSON
type JSONResponseWriter struct {
	http.ResponseWriter
	Buffer *bytes.Buffer
}

// Write captures the written response
func (w *JSONResponseWriter) Write(b []byte) (int, error) {
	return w.Buffer.Write(b)
}

// JSONFormatterMiddleware ensures that all JSON responses are properly formatted
func JSONFormatterMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Initialize buffer and wrapper
		buffer := &bytes.Buffer{}
		wrapper := &JSONResponseWriter{
			ResponseWriter: w,
			Buffer:         buffer,
		}

		// Execute the handler with our wrapper
		next.ServeHTTP(wrapper, r)

		// Check if content type is JSON
		contentType := w.Header().Get("Content-Type")
		isJSON := contentType == "application/json" || contentType == "" // If empty, we assume JSON

		if isJSON && buffer.Len() > 0 {
			var data interface{}
			
			// Try to decode JSON from buffer
			if err := json.Unmarshal(buffer.Bytes(), &data); err == nil {
				// Reencode with pretty formatting
				formattedJSON, err := json.MarshalIndent(data, "", "    ")
				if err == nil {
					// Set Content-Type if not already defined
					if contentType == "" {
						w.Header().Set("Content-Type", "application/json")
					}
					// Write content length header
					w.Header().Set("Content-Length", strconv.Itoa(len(formattedJSON)))
					
					// Write formatted JSON
					w.Write(formattedJSON)
					return
				}
			}
		}
		
		// If not JSON or can't format, write original buffer
		w.Write(buffer.Bytes())
	})
}

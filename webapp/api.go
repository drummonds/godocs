package webapp

import (
	"github.com/maxence-charriere/go-app/v10/pkg/app"
)

// GetAPIBaseURL returns the configured API base URL
// It reads from window.goEDMSConfig.apiURL if available,
// otherwise falls back to empty string (relative URLs)
func GetAPIBaseURL() string {
	// Check if config is available in browser
	if !app.IsClient {
		return "" // Server-side rendering - use relative URLs
	}

	// Try to get API URL from global config
	config := app.Window().Get("goEDMSConfig")
	if config.Truthy() {
		apiURL := config.Get("apiURL")
		if apiURL.Truthy() {
			url := apiURL.String()
			// Ensure no trailing slash
			if len(url) > 0 && url[len(url)-1] == '/' {
				return url[:len(url)-1]
			}
			return url
		}
	}

	// Fallback to relative URLs (same origin)
	return ""
}

// BuildAPIURL constructs a full API URL from a path
// Example: BuildAPIURL("/api/documents/latest") -> "http://backend:8000/api/documents/latest"
// or just "/api/documents/latest" if using relative URLs
func BuildAPIURL(path string) string {
	baseURL := GetAPIBaseURL()
	if baseURL == "" {
		return path // Relative URL
	}
	return baseURL + path
}

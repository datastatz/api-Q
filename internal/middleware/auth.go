package middleware

import (
	"apiq/internal/database"
	"encoding/json"
	"net/http"
)

// APIKeyAuth middleware controleert of de API key geldig is
func APIKeyAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Haal API key op uit de request header
		apiKey := r.Header.Get("X-API-Key")

		// Controleer of API key is meegegeven
		if apiKey == "" {
			w.WriteHeader(http.StatusUnauthorized)
			response := map[string]string{
				"error": "API key required. Add 'X-API-Key' header to your request.",
			}
			json.NewEncoder(w).Encode(response)
			return
		}

		// Controleer of API key bestaat in database
		var apiKeyRecord database.APIKey
		result := database.DB.Where("api_key = ? AND is_active = ?", apiKey, true).First(&apiKeyRecord)

		if result.Error != nil {
			w.WriteHeader(http.StatusUnauthorized)
			response := map[string]string{
				"error": "Invalid API key",
			}
			json.NewEncoder(w).Encode(response)
			return
		}

		// API key is geldig, ga door naar de volgende functie
		next(w, r)
	}
}

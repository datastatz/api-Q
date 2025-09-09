package handlers

import (
	"apiq/internal/database"
	"encoding/json"
	"net/http"
	"time"
)

// Admin analytics - alle bedrijven
func AdminAnalytics(w http.ResponseWriter, r *http.Request) {
	// Haal alle request logs op van de afgelopen 12 maanden
	var logs []database.RequestLog
	twelveMonthsAgo := time.Now().AddDate(0, -12, 0)

	database.DB.Where("timestamp >= ?", twelveMonthsAgo).Find(&logs)

	// Groepeer per bedrijf per maand
	companyStats := make(map[string]map[string]interface{})

	for _, log := range logs {
		// Haal bedrijfsnaam op
		var apiKey database.APIKey
		database.DB.Where("api_key = ?", log.APIKey).First(&apiKey)

		if apiKey.CompanyName == "" {
			continue
		}

		// Maak maand string (2024-01)
		month := log.Timestamp.Format("2006-01")

		if companyStats[apiKey.CompanyName] == nil {
			companyStats[apiKey.CompanyName] = make(map[string]interface{})
		}

		if companyStats[apiKey.CompanyName][month] == nil {
			companyStats[apiKey.CompanyName][month] = map[string]interface{}{
				"requests": 0,
				"cost":     0.0,
			}
		}

		// Update stats
		monthData := companyStats[apiKey.CompanyName][month].(map[string]interface{})
		monthData["requests"] = monthData["requests"].(int) + 1
		monthData["cost"] = monthData["cost"].(float64) + log.Cost
	}

	response := map[string]interface{}{
		"analytics": companyStats,
		"period":    "12 months",
	}
	json.NewEncoder(w).Encode(response)
}

// Company analytics - eigen data
func CompanyAnalytics(w http.ResponseWriter, r *http.Request) {
	// Haal API key op uit header
	apiKey := r.Header.Get("X-API-Key")
	if apiKey == "" {
		http.Error(w, "API key required", http.StatusUnauthorized)
		return
	}

	// Haal bedrijfsnaam op
	var apiKeyRecord database.APIKey
	result := database.DB.Where("api_key = ?", apiKey).First(&apiKeyRecord)
	if result.Error != nil {
		http.Error(w, "Invalid API key", http.StatusUnauthorized)
		return
	}

	// Haal eigen request logs op van de afgelopen 12 maanden
	var logs []database.RequestLog
	twelveMonthsAgo := time.Now().AddDate(0, -12, 0)

	database.DB.Where("api_key = ? AND timestamp >= ?", apiKey, twelveMonthsAgo).Find(&logs)

	// Groepeer per maand
	monthlyStats := make(map[string]interface{})

	for _, log := range logs {
		month := log.Timestamp.Format("2006-01")

		if monthlyStats[month] == nil {
			monthlyStats[month] = map[string]interface{}{
				"requests": 0,
				"cost":     0.0,
			}
		}

		monthData := monthlyStats[month].(map[string]interface{})
		monthData["requests"] = monthData["requests"].(int) + 1
		monthData["cost"] = monthData["cost"].(float64) + log.Cost
	}

	response := map[string]interface{}{
		"company_name": apiKeyRecord.CompanyName,
		"analytics":    monthlyStats,
		"period":       "12 months",
	}
	json.NewEncoder(w).Encode(response)
}

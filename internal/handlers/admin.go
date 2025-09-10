package handlers

import (
	"apiq/internal/database"
	"encoding/json"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
	"time"
)

// Admin credentials (in productie: in database of environment)
const ADMIN_USERNAME = "admin"
const ADMIN_PASSWORD = "admin123"     // Verander dit!
const JWT_SECRET = "stellarisdebeste" // Verander dit!

// Admin login endpoint
func AdminLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var loginData struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&loginData); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Controleer credentials
	if loginData.Username != ADMIN_USERNAME || loginData.Password != ADMIN_PASSWORD {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Maak JWT token aan (24 uur geldig)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": loginData.Username,
		"exp":      time.Now().Add(time.Hour * 24).Unix(), // 24 uur geldig
	})

	tokenString, err := token.SignedString([]byte(JWT_SECRET))
	if err != nil {
		http.Error(w, "Error creating token", http.StatusInternalServerError)
		return
	}

	// Stuur token terug
	response := map[string]string{
		"token": tokenString,
	}
	json.NewEncoder(w).Encode(response)
}

// Alle bedrijven ophalen
func GetAllCompanies(w http.ResponseWriter, r *http.Request) {
	var companies []database.APIKey
	database.DB.Find(&companies)

	response := map[string]interface{}{
		"companies": companies,
		"total":     len(companies),
	}
	json.NewEncoder(w).Encode(response)
}

// Nieuwe API key aanmaken
func CreateAPIKey(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var requestData struct {
		CompanyName string `json:"company_name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Maak nieuwe API key
	apiKey, err := database.CreateNewAPIKey(requestData.CompanyName)
	if err != nil {
		http.Error(w, "Error creating API key", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"api_key":      apiKey.APIKey,
		"company_name": apiKey.CompanyName,
		"created_at":   apiKey.CreatedAt,
	}
	json.NewEncoder(w).Encode(response)
}

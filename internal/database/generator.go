package database

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

// APIKEY GENERATOR

// GenerateAPIKey maakt een nieuwe unieke API key
func GenerateAPIKey() string {
	// Genereer 16 random bytes
	bytes := make([]byte, 16)
	rand.Read(bytes)

	// Converteer naar hex string
	randomString := hex.EncodeToString(bytes)

	// Voeg "ak_" prefix toe
	return fmt.Sprintf("ak_%s", randomString)
}

// CreateNewAPIKey maakt een nieuwe API key aan voor een bedrijf
func CreateNewAPIKey(companyName string) (*APIKey, error) {
	apiKey := &APIKey{
		APIKey:      GenerateAPIKey(),
		CompanyName: companyName,
		IsActive:    true,
	}

	// Sla op in database
	result := DB.Create(apiKey)
	if result.Error != nil {
		return nil, result.Error
	}

	return apiKey, nil
}

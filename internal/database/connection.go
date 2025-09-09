package database

import (
	"log"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

// InitDatabase initialiseert de database connectie
func InitDatabase() {
	var err error

	// Maak verbinding met SQLite database
	DB, err = gorm.Open(sqlite.Open("apiq.db"), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Maak tabellen aan (als ze nog niet bestaan)
	err = DB.AutoMigrate(&APIKey{}, &RequestLog{})
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	log.Println("Database connected and migrated successfully")
}

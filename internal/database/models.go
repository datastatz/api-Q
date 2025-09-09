package database

import (
	"time"
)

// API Keys met bedrijfsnamen (Voor ons)
type APIKey struct {
	ID          uint      `gorm:"primaryKey"`
	APIKey      string    `gorm:"uniqueIndex;not null"`
	CompanyName string    `gorm:"not null"`
	CreatedAt   time.Time `gorm:"autoCreateTime"`
	IsActive    bool      `gorm:"default:true"`
}

// Request logs - alleen wat we nodig hebben
type RequestLog struct {
	ID        uint      `gorm:"primaryKey"`
	APIKey    string    `gorm:"not null"`       // Welke bedrijf
	Timestamp time.Time `gorm:"autoCreateTime"` // Wanneer
	Cost      float64   `gorm:"not null"`       // Kosten per request
}

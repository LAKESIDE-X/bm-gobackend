package models

import "gorm.io/gorm"

// The blueprint for a Wellness Hub Article
type Article struct {
	gorm.Model
	Title    string `json:"title" gorm:"not null"`
	Category string `json:"category" gorm:"not null"` // e.g., "IMMUNITY", "SKINCARE"
	Content  string `json:"content" gorm:"type:text;not null"`
	ImageURL string `json:"imageUrl"`
}

// The blueprint for Newsletter Subscribers
type Subscriber struct {
	gorm.Model
	Email string `json:"email" gorm:"unique;not null"`
}

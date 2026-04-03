// models/review.go
package models

import (
	"time"
)

type Review struct {
	ID        uint `gorm:"primaryKey" json:"id"`
	ProductID uint `gorm:"not null" json:"productId"`
	UserID    uint `gorm:"not null" json:"userId"`

	// Preload the User to display their name on the frontend!
	User User `gorm:"foreignKey:UserID" json:"user"`

	Rating    int       `gorm:"not null" json:"rating"` // e.g., 1 to 5
	Comment   string    `gorm:"type:text;not null" json:"comment"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

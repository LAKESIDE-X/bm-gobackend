// models/cart.go
package models

import (
	"time"
)

type CartItem struct {
	ID        uint `gorm:"primaryKey" json:"id"`
	UserID    uint `gorm:"not null" json:"userId"`
	ProductID uint `gorm:"not null" json:"productId"`

	// This tells GORM to load the full Product details when we fetch the cart!
	Product Product `gorm:"foreignKey:ProductID" json:"product"`

	Quantity  int       `gorm:"not null;default:1" json:"quantity"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

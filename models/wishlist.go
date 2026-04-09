package models

import "gorm.io/gorm"

type Wishlist struct {
	gorm.Model
	UserID    uint    `json:"userId"`
	ProductID uint    `json:"productId"`
	Product   Product `gorm:"foreignKey:ProductID" json:"product"`
}

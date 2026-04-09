package models

import "gorm.io/gorm"

type Review struct {
	gorm.Model
	UserID    uint   `json:"userId"`
	ProductID uint   `json:"productId"`
	Rating    int    `json:"rating"` // 1 to 5 stars
	Comment   string `json:"comment"`
	User      User   `gorm:"foreignKey:UserID" json:"user"` // This allows us to show the User's name on the review
}

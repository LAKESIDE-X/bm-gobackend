package models

import (
	"time"
	"gorm.io/gorm"
)

type User struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	FirstName string         `gorm:"size:100;not null" json:"firstName"`
	LastName  string         `gorm:"size:100;not null" json:"lastName"`
	Email     string         `gorm:"size:100;not null;unique" json:"email"`
	Phone     string         `gorm:"size:20" json:"phone"`
	Password  string         `gorm:"not null" json:"-"` // The "-" hides the password from JSON responses!
	Role      string         `gorm:"type:varchar(20);default:'USER'" json:"role"` // 'USER' or 'ADMIN'
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
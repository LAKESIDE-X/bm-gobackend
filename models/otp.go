package models

import (
	"time"
)

type OTP struct {
	ID        uint      `gorm:"primaryKey"`
	Email     string    `gorm:"size:100;not null;index"`
	Code      string    `gorm:"size:6;not null"`
	ExpiresAt time.Time `gorm:"not null"`
	CreatedAt time.Time
}

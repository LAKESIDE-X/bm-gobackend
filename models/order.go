// models/order.go
package models

import (
	"time"
)

type Order struct {
	ID              uint        `gorm:"primaryKey" json:"id"`
	UserID          uint        `gorm:"not null" json:"userId"`
	User            User        `gorm:"foreignKey:UserID" json:"user"`
	TotalAmount     float64     `gorm:"not null" json:"totalAmount"`
	ShippingAddress string      `gorm:"type:text;not null" json:"shippingAddress"`
	PaymentMethod   string      `gorm:"size:50;not null" json:"paymentMethod"`
	Status          string      `gorm:"size:50;default:'PENDING'" json:"status"` // PENDING, PROCESSING, SHIPPED, DELIVERED, CANCELLED
	OrderItems      []OrderItem `gorm:"foreignKey:OrderID" json:"items"`
	CreatedAt       time.Time   `json:"createdAt"`
	UpdatedAt       time.Time   `json:"updatedAt"`
}

type OrderItem struct {
	ID              uint    `gorm:"primaryKey" json:"id"`
	OrderID         uint    `gorm:"not null" json:"orderId"`
	ProductID       uint    `gorm:"not null" json:"productId"`
	Product         Product `gorm:"foreignKey:ProductID" json:"product"`
	Quantity        int     `gorm:"not null" json:"quantity"`
	PriceAtPurchase float64 `gorm:"not null" json:"priceAtPurchase"` // Locks in the price in case the admin changes it later!
}

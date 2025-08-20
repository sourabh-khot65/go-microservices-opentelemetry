package models

import "time"

type OrderItem struct {
	ProductID string `json:"product_id" db:"product_id" binding:"required"`
	Quantity  int    `json:"quantity" db:"quantity" binding:"required,gt=0"`
}

type Order struct {
	ID        string    `json:"id" db:"id"`
	ProductID string    `json:"product_id" db:"product_id" binding:"required,uuid"`
	Quantity  int       `json:"quantity" db:"quantity" binding:"required,gt=0"`
	Status    string    `json:"status" db:"status"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type CreateOrderRequest struct {
	ProductID string `json:"product_id" binding:"required,uuid"`
	Quantity  int    `json:"quantity" binding:"required,gt=0"`
}

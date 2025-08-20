package repository

import (
	"context"
	"database/sql"
	"order-service/internal/models"
)

type OrderRepository struct {
	DB *sql.DB
}

func NewOrderRepository(db *sql.DB) *OrderRepository {
	return &OrderRepository{DB: db}
}

func (r *OrderRepository) Create(ctx context.Context, order *models.Order) error {
	_, err := r.DB.ExecContext(ctx, `INSERT INTO orders (id, items) VALUES ($1, $2)`, order.ID, order.Items)
	return err
}

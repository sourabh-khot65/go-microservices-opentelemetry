package repository

import (
	"context"
	"database/sql"
	"order-service/internal/models"
	"time"

	"github.com/google/uuid"
)

type OrderRepository struct {
	DB *sql.DB
}

func NewOrderRepository(db *sql.DB) *OrderRepository {
	return &OrderRepository{DB: db}
}

func (r *OrderRepository) Create(ctx context.Context, req *models.CreateOrderRequest) (*models.Order, error) {
	order := &models.Order{
		ID:        uuid.New().String(),
		ProductID: req.ProductID,
		Quantity:  req.Quantity,
		Status:    "pending",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	_, err := r.DB.ExecContext(ctx,
		`INSERT INTO orders (id, product_id, quantity, status, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6)`,
		order.ID, order.ProductID, order.Quantity, order.Status, order.CreatedAt, order.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return order, nil
}

func (r *OrderRepository) GetByID(ctx context.Context, id string) (*models.Order, error) {
	order := &models.Order{}
	err := r.DB.QueryRowContext(ctx,
		`SELECT id, product_id, quantity, status, created_at, updated_at FROM orders WHERE id = $1`,
		id,
	).Scan(&order.ID, &order.ProductID, &order.Quantity, &order.Status, &order.CreatedAt, &order.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return order, nil
}

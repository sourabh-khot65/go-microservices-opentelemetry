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

func (r *OrderRepository) GetAll(ctx context.Context, page, pageSize int) ([]*models.Order, int64, error) {
	offset := (page - 1) * pageSize

	// Get total count
	var total int64
	err := r.DB.QueryRowContext(ctx, `SELECT COUNT(*) FROM orders`).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Get orders with pagination
	rows, err := r.DB.QueryContext(ctx,
		`SELECT id, product_id, quantity, status, created_at, updated_at 
		 FROM orders 
		 ORDER BY created_at DESC 
		 LIMIT $1 OFFSET $2`,
		pageSize, offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var orders []*models.Order
	for rows.Next() {
		order := &models.Order{}
		err := rows.Scan(&order.ID, &order.ProductID, &order.Quantity, &order.Status, &order.CreatedAt, &order.UpdatedAt)
		if err != nil {
			return nil, 0, err
		}
		orders = append(orders, order)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, err
	}

	return orders, total, nil
}

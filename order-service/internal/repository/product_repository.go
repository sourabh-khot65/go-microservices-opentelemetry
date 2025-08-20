package repository

import (
	"context"
	"database/sql"
	"order-service/internal/models"
	"time"

	"github.com/google/uuid"
)

type ProductRepository struct {
	DB *sql.DB
}

func NewProductRepository(db *sql.DB) *ProductRepository {
	return &ProductRepository{DB: db}
}

func (r *ProductRepository) Create(ctx context.Context, req *models.CreateProductRequest) (*models.Product, error) {
	product := &models.Product{
		ID:          uuid.New().String(),
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Stock:       req.Stock,
		CreatedAt:   time.Now(),
	}

	_, err := r.DB.ExecContext(ctx,
		`INSERT INTO products (id, name, description, price, stock, created_at) VALUES ($1, $2, $3, $4, $5, $6)`,
		product.ID, product.Name, product.Description, product.Price, product.Stock, product.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return product, nil
}

func (r *ProductRepository) GetByID(ctx context.Context, id string) (*models.Product, error) {
	product := &models.Product{}
	err := r.DB.QueryRowContext(ctx,
		`SELECT id, name, description, price, stock, created_at FROM products WHERE id = $1`,
		id,
	).Scan(&product.ID, &product.Name, &product.Description, &product.Price, &product.Stock, &product.CreatedAt)
	
	if err != nil {
		return nil, err
	}
	
	return product, nil
}

func (r *ProductRepository) GetAll(ctx context.Context) ([]*models.Product, error) {
	rows, err := r.DB.QueryContext(ctx,
		`SELECT id, name, description, price, stock, created_at FROM products ORDER BY created_at DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []*models.Product
	for rows.Next() {
		product := &models.Product{}
		err := rows.Scan(&product.ID, &product.Name, &product.Description, &product.Price, &product.Stock, &product.CreatedAt)
		if err != nil {
			return nil, err
		}
		products = append(products, product)
	}

	return products, rows.Err()
}

func (r *ProductRepository) Exists(ctx context.Context, id string) (bool, error) {
	var exists bool
	err := r.DB.QueryRowContext(ctx,
		`SELECT EXISTS(SELECT 1 FROM products WHERE id = $1)`,
		id,
	).Scan(&exists)
	
	return exists, err
}
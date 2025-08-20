package service

import (
	"context"
	"order-service/internal/models"
	"order-service/internal/repository"
)

type OrderService struct {
	Repo *repository.OrderRepository
}

func NewOrderService(repo *repository.OrderRepository) *OrderService {
	return &OrderService{Repo: repo}
}

func (s *OrderService) CreateOrder(ctx context.Context, req *models.CreateOrderRequest) (*models.Order, error) {
	return s.Repo.Create(ctx, req)
}

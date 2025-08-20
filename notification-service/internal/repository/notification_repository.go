package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"notification-service/internal/models"
	"time"

	"github.com/google/uuid"
)

type NotificationRepository struct {
	DB *sql.DB
}

func NewNotificationRepository(db *sql.DB) *NotificationRepository {
	return &NotificationRepository{DB: db}
}

func (r *NotificationRepository) Create(ctx context.Context, req *models.NotificationRequest) (*models.Notification, error) {
	notification := &models.Notification{
		ID:        uuid.New().String(),
		Message:   req.Message,
		OrderID:   req.OrderID,
		CreatedAt: time.Now(),
	}

	// Convert trace data to JSON for storage
	traceDataJSON, err := json.Marshal(notification.TraceData)
	if err != nil {
		return nil, err
	}

	_, err = r.DB.ExecContext(ctx,
		`INSERT INTO notifications (id, message, order_id, trace_data, created_at) VALUES ($1, $2, $3, $4, $5)`,
		notification.ID, notification.Message, notification.OrderID, traceDataJSON, notification.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return notification, nil
}

func (r *NotificationRepository) GetByID(ctx context.Context, id string) (*models.Notification, error) {
	notification := &models.Notification{}
	var traceDataJSON []byte

	err := r.DB.QueryRowContext(ctx,
		`SELECT id, message, order_id, trace_data, created_at FROM notifications WHERE id = $1`,
		id,
	).Scan(&notification.ID, &notification.Message, &notification.OrderID, &traceDataJSON, &notification.CreatedAt)

	if err != nil {
		return nil, err
	}

	// Parse JSON trace data
	if len(traceDataJSON) > 0 {
		err = json.Unmarshal(traceDataJSON, &notification.TraceData)
		if err != nil {
			return nil, err
		}
	}

	return notification, nil
}

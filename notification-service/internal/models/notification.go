package models

import "time"

type Notification struct {
	ID        string                 `json:"id"`
	Message   string                 `json:"message"`
	OrderID   string                 `json:"order_id"`
	TraceData map[string]interface{} `json:"trace_data"`
	CreatedAt time.Time              `json:"created_at"`
}

type NotificationRequest struct {
	Message string `json:"message" binding:"required"`
	OrderID string `json:"order_id" binding:"required"`
}

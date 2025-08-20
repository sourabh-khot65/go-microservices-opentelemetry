package api

import (
	"log/slog"
	"net/http"
	"notification-service/internal/models"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type NotificationHandler struct {
	tracer trace.Tracer
}

func NewNotificationHandler() *NotificationHandler {
	return &NotificationHandler{
		tracer: otel.Tracer("notification-handler"),
	}
}

func (h *NotificationHandler) RegisterRoutes(router *gin.Engine) {
	api := router.Group("/api")
	{
		api.POST("/notifications", h.CreateNotification)
		api.GET("/notifications/:id", h.GetNotification)
	}
}

func (h *NotificationHandler) CreateNotification(c *gin.Context) {
	ctx, span := h.tracer.Start(c.Request.Context(), "create_notification")
	defer span.End()

	requestID := c.GetString("request_id")

	var req models.NotificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		slog.ErrorContext(ctx, "Invalid notification request",
			"error", err.Error(),
			"request_id", requestID,
		)
		span.RecordError(err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request payload",
			"request_id": requestID,
		})
		return
	}

	// Extract trace information from the request
	spanContext := trace.SpanContextFromContext(ctx)
	
	notification := &models.Notification{
		ID:        uuid.New().String(),
		Message:   req.Message,
		OrderID:   req.OrderID,
		TraceData: make(map[string]interface{}),
		CreatedAt: time.Now(),
	}

	// Add current trace context to notification
	notification.TraceData["notification_trace_id"] = spanContext.TraceID().String()
	notification.TraceData["notification_span_id"] = spanContext.SpanID().String()
	notification.TraceData["notification_service"] = "notification-service"

	// Log the notification data - trace context automatically included by OpenTelemetry
	slog.InfoContext(ctx, "📧 NOTIFICATION RECEIVED AND PROCESSED",
		"notification_id", notification.ID,
		"order_id", notification.OrderID,
		"message", notification.Message,
		"distributed_trace_flow", "order-service → notification-service",
		"request_id", requestID,
	)

	span.SetAttributes(
		attribute.String("notification.id", notification.ID),
		attribute.String("notification.order_id", notification.OrderID),
		attribute.String("notification.message", notification.Message),
		attribute.String("distributed_trace.status", "success"),
		attribute.String("distributed_trace.flow", "order-service->notification-service"),
	)

	// Trace context is automatically propagated via HTTP headers by OpenTelemetry

	c.JSON(http.StatusCreated, notification)
}

func (h *NotificationHandler) GetNotification(c *gin.Context) {
	ctx, span := h.tracer.Start(c.Request.Context(), "get_notification")
	defer span.End()

	requestID := c.GetString("request_id")
	notificationID := c.Param("id")

	span.SetAttributes(
		attribute.String("notification.id", notificationID),
		attribute.String("method", "GET"),
	)

	slog.InfoContext(ctx, "📱 GET NOTIFICATION REQUEST",
		"notification_id", notificationID,
		"request_id", requestID,
		"operation", "retrieve_notification",
		"service", "notification-service",
	)

	// Simple response - no database needed, just log the request
	response := gin.H{
		"id":         notificationID,
		"message":    "Notification service is operational - check logs for distributed tracing data",
		"service":    "notification-service",
		"status":     "active",
		"request_id": requestID,
		"note":       "Notifications are logged, not stored",
	}

	c.JSON(http.StatusOK, response)
}

package api

import (
	"log/slog"
	"net/http"
	"notification-service/internal/observability"

	"github.com/gin-gonic/gin"
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
	router.GET("/notifications", h.HandleNotification)
}

func (h *NotificationHandler) HandleNotification(c *gin.Context) {
	ctx, span := h.tracer.Start(c.Request.Context(), "handle_notification")
	defer span.End()

	span.SetAttributes(
		attribute.String("handler", "notification"),
		attribute.String("method", "GET"),
	)

	slog.InfoContext(ctx, "Processing notification request",
		"path", c.Request.URL.Path,
		"method", c.Request.Method,
	)

	response := gin.H{
		"message":    "Notification endpoint is working",
		"service":    "notification-service",
		"version":    "1.0.0",
		"request_id": observability.GetRequestID(ctx),
	}

	c.JSON(http.StatusOK, response)
}

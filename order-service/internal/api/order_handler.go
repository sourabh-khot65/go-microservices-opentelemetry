package api

import (
	"log/slog"
	"net/http"
	"order-service/internal/models"
	"order-service/internal/observability"
	"order-service/internal/service"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type OrderHandler struct {
	Service *service.OrderService
	tracer  trace.Tracer
}

func NewOrderHandler(svc *service.OrderService) *OrderHandler {
	return &OrderHandler{
		Service: svc,
		tracer:  otel.Tracer("order-handler"),
	}
}

func (h *OrderHandler) RegisterRoutes(r *gin.Engine) {
	r.POST("/orders", h.CreateOrder)
}

func (h *OrderHandler) CreateOrder(c *gin.Context) {
	ctx, span := h.tracer.Start(c.Request.Context(), "create_order")
	defer span.End()

	var order models.Order
	if err := c.ShouldBindJSON(&order); err != nil {
		span.SetAttributes(attribute.Bool("error", true))
		slog.ErrorContext(ctx, "Invalid order data", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error":      "Invalid request data",
			"message":    err.Error(),
			"request_id": observability.GetRequestID(ctx),
		})
		return
	}

	span.SetAttributes(
		attribute.String("order.id", order.ID),
		attribute.Int("order.items_count", len(order.Items)),
	)

	if err := h.Service.CreateOrder(ctx, &order); err != nil {
		span.SetAttributes(attribute.Bool("error", true))
		slog.ErrorContext(ctx, "Failed to create order", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":      "Failed to create order",
			"message":    "An internal error occurred",
			"request_id": observability.GetRequestID(ctx),
		})
		return
	}

	slog.InfoContext(ctx, "Order created successfully",
		"order_id", order.ID,
		"items_count", len(order.Items),
	)

	c.JSON(http.StatusCreated, gin.H{
		"order":      order,
		"message":    "Order created successfully",
		"request_id": observability.GetRequestID(ctx),
	})
}

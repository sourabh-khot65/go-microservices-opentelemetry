package api

import (
	"database/sql"
	"log/slog"
	"net/http"
	"order-service/internal/models"
	"order-service/internal/repository"
	"order-service/internal/services"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type OrderHandler struct {
	orderRepo           *repository.OrderRepository
	productRepo         *repository.ProductRepository
	notificationService *services.NotificationService
	tracer              trace.Tracer
}

func NewOrderHandler(orderRepo *repository.OrderRepository, productRepo *repository.ProductRepository, notificationService *services.NotificationService) *OrderHandler {
	return &OrderHandler{
		orderRepo:           orderRepo,
		productRepo:         productRepo,
		notificationService: notificationService,
		tracer:              otel.Tracer("order-handler"),
	}
}

func (h *OrderHandler) CreateOrder(c *gin.Context) {
	ctx, span := h.tracer.Start(c.Request.Context(), "create_order")
	defer span.End()

	requestID := c.GetString("request_id")

	var req models.CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		slog.ErrorContext(ctx, "Invalid order request",
			"error", err.Error(),
			"request_id", requestID,
		)
		span.RecordError(err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error":      "Invalid request payload",
			"request_id": requestID,
		})
		return
	}

	// Validate that product exists
	exists, err := h.productRepo.Exists(ctx, req.ProductID)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to check product existence",
			"error", err.Error(),
			"product_id", req.ProductID,
			"request_id", requestID,
		)
		span.RecordError(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":      "Failed to validate product",
			"request_id": requestID,
		})
		return
	}
	if !exists {
		slog.WarnContext(ctx, "Product not found",
			"product_id", req.ProductID,
			"request_id", requestID,
		)
		c.JSON(http.StatusBadRequest, gin.H{
			"error":      "Product not found: " + req.ProductID,
			"request_id": requestID,
		})
		return
	}

	// Create the order
	order, err := h.orderRepo.Create(ctx, &req)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to create order",
			"error", err.Error(),
			"request_id", requestID,
		)
		span.RecordError(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":      "Failed to create order",
			"request_id": requestID,
		})
		return
	}

	// Send notification with distributed tracing
	if err := h.notificationService.SendOrderCreatedNotification(ctx, order.ID, order.ProductID, order.Quantity); err != nil {
		slog.ErrorContext(ctx, "Failed to send notification",
			"error", err.Error(),
			"order_id", order.ID,
			"request_id", requestID,
		)
		// Don't fail the order creation, just log the error
		span.RecordError(err)
	}

	slog.InfoContext(ctx, "Order created successfully",
		"order_id", order.ID,
		"product_id", order.ProductID,
		"quantity", order.Quantity,
		"request_id", requestID,
	)

	span.SetAttributes(
		attribute.String("order.id", order.ID),
		attribute.String("order.product_id", order.ProductID),
		attribute.Int("order.quantity", order.Quantity),
		attribute.String("order.status", order.Status),
	)

	c.JSON(http.StatusCreated, order)
}

func (h *OrderHandler) GetOrder(c *gin.Context) {
	ctx, span := h.tracer.Start(c.Request.Context(), "get_order")
	defer span.End()

	requestID := c.GetString("request_id")
	orderID := c.Param("id")

	order, err := h.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		if err == sql.ErrNoRows {
			slog.WarnContext(ctx, "Order not found",
				"order_id", orderID,
				"request_id", requestID,
			)
			c.JSON(http.StatusNotFound, gin.H{
				"error":      "Order not found",
				"request_id": requestID,
			})
			return
		}

		slog.ErrorContext(ctx, "Failed to get order",
			"error", err.Error(),
			"order_id", orderID,
			"request_id", requestID,
		)
		span.RecordError(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":      "Failed to get order",
			"request_id": requestID,
		})
		return
	}

	span.SetAttributes(
		attribute.String("order.id", order.ID),
		attribute.String("order.status", order.Status),
	)

	c.JSON(http.StatusOK, order)
}

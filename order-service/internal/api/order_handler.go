package api

import (
	"database/sql"
	"log/slog"
	"math"
	"net/http"
	"order-service/internal/models"
	"order-service/internal/repository"
	"order-service/internal/services"
	"strconv"

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

	// Get order
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

	// Get product details
	product, err := h.productRepo.GetByID(ctx, order.ProductID)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to get product details",
			"error", err.Error(),
			"product_id", order.ProductID,
			"request_id", requestID,
		)
		span.RecordError(err)
		// Still return order even if product fetch fails
		c.JSON(http.StatusOK, gin.H{
			"order":      order,
			"product":    nil,
			"request_id": requestID,
		})
		return
	}

	span.SetAttributes(
		attribute.String("order.id", order.ID),
		attribute.String("order.status", order.Status),
		attribute.String("product.id", product.ID),
		attribute.String("product.name", product.Name),
	)

	c.JSON(http.StatusOK, gin.H{
		"order":      order,
		"product":    product,
		"request_id": requestID,
	})
}

// GetAllOrders retrieves all orders with pagination and product details
func (h *OrderHandler) GetAllOrders(c *gin.Context) {
	ctx, span := h.tracer.Start(c.Request.Context(), "get_all_orders")
	defer span.End()

	requestID := c.GetString("request_id")

	// Parse pagination parameters
	page := 1
	pageSize := 10
	if p := c.Query("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}
	if ps := c.Query("page_size"); ps != "" {
		if parsed, err := strconv.Atoi(ps); err == nil && parsed > 0 && parsed <= 100 {
			pageSize = parsed
		}
	}

	span.SetAttributes(
		attribute.Int("pagination.page", page),
		attribute.Int("pagination.page_size", pageSize),
	)

	// Get orders from repository
	orders, total, err := h.orderRepo.GetAll(ctx, page, pageSize)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to get orders",
			"error", err.Error(),
			"request_id", requestID,
		)
		span.RecordError(err)
		response := models.CreateErrorResponse("Failed to get orders", "INTERNAL_ERROR", requestID)
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	// Calculate pagination metadata
	totalPages := int(math.Ceil(float64(total) / float64(pageSize)))
	pagination := &models.Pagination{
		Page:       page,
		PageSize:   pageSize,
		Total:      total,
		TotalPages: totalPages,
	}

	// Get product details for each order
	ordersWithProducts := make([]*models.OrderWithProduct, 0, len(orders))
	for _, order := range orders {
		orderWithProduct := &models.OrderWithProduct{
			Order: order,
		}

		// Try to get product details
		if product, err := h.productRepo.GetByID(ctx, order.ProductID); err == nil {
			orderWithProduct.Product = product
		} else {
			slog.WarnContext(ctx, "Failed to get product details for order",
				"error", err.Error(),
				"order_id", order.ID,
				"product_id", order.ProductID,
				"request_id", requestID,
			)
			// Continue without product details
		}

		ordersWithProducts = append(ordersWithProducts, orderWithProduct)
	}

	slog.InfoContext(ctx, "GetAllOrders completed successfully",
		"page", page,
		"page_size", pageSize,
		"total", total,
		"returned_count", len(ordersWithProducts),
		"request_id", requestID,
	)

	span.SetAttributes(
		attribute.Int64("orders.total", total),
		attribute.Int("orders.returned", len(ordersWithProducts)),
		attribute.String("endpoint.status", "success"),
	)

	response := models.CreatePaginatedResponse(ordersWithProducts, pagination, requestID)
	c.JSON(http.StatusOK, response)
}

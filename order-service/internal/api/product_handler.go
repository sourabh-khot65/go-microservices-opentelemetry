package api

import (
	"database/sql"
	"log/slog"
	"net/http"
	"order-service/internal/models"
	"order-service/internal/repository"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type ProductHandler struct {
	productRepo *repository.ProductRepository
	tracer      trace.Tracer
}

func NewProductHandler(productRepo *repository.ProductRepository) *ProductHandler {
	return &ProductHandler{
		productRepo: productRepo,
		tracer:      otel.Tracer("product-handler"),
	}
}

func (h *ProductHandler) CreateProduct(c *gin.Context) {
	ctx, span := h.tracer.Start(c.Request.Context(), "create_product")
	defer span.End()

	requestID := c.GetString("request_id")

	var req models.CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		slog.ErrorContext(ctx, "Invalid request payload",
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

	product, err := h.productRepo.Create(ctx, &req)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to create product",
			"error", err.Error(),
			"request_id", requestID,
		)
		span.RecordError(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":      "Failed to create product",
			"request_id": requestID,
		})
		return
	}

	slog.InfoContext(ctx, "Product created successfully",
		"product_id", product.ID,
		"product_name", product.Name,
		"request_id", requestID,
	)

	span.SetAttributes(
		attribute.String("product.id", product.ID),
		attribute.String("product.name", product.Name),
	)

	c.JSON(http.StatusCreated, product)
}

func (h *ProductHandler) GetProduct(c *gin.Context) {
	ctx, span := h.tracer.Start(c.Request.Context(), "get_product")
	defer span.End()

	requestID := c.GetString("request_id")
	productID := c.Param("id")

	product, err := h.productRepo.GetByID(ctx, productID)
	if err != nil {
		if err == sql.ErrNoRows {
			slog.WarnContext(ctx, "Product not found",
				"product_id", productID,
				"request_id", requestID,
			)
			c.JSON(http.StatusNotFound, gin.H{
				"error":      "Product not found",
				"request_id": requestID,
			})
			return
		}

		slog.ErrorContext(ctx, "Failed to get product",
			"error", err.Error(),
			"product_id", productID,
			"request_id", requestID,
		)
		span.RecordError(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":      "Failed to get product",
			"request_id": requestID,
		})
		return
	}

	span.SetAttributes(
		attribute.String("product.id", product.ID),
		attribute.String("product.name", product.Name),
	)

	c.JSON(http.StatusOK, product)
}

func (h *ProductHandler) GetAllProducts(c *gin.Context) {
	ctx, span := h.tracer.Start(c.Request.Context(), "get_all_products")
	defer span.End()

	requestID := c.GetString("request_id")

	products, err := h.productRepo.GetAll(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to get products",
			"error", err.Error(),
			"request_id", requestID,
		)
		span.RecordError(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":      "Failed to get products",
			"request_id": requestID,
		})
		return
	}

	span.SetAttributes(
		attribute.Int64("products.count", int64(len(products))),
	)

	c.JSON(http.StatusOK, gin.H{
		"products":   products,
		"count":      len(products),
		"request_id": requestID,
	})
}

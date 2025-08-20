package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

type NotificationRequest struct {
	Message string `json:"message"`
	OrderID string `json:"order_id"`
}

type NotificationService struct {
	baseURL    string
	httpClient *http.Client
	tracer     trace.Tracer
}

func NewNotificationService(baseURL string) *NotificationService {
	return &NotificationService{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout:   30 * time.Second,
			Transport: otelhttp.NewTransport(http.DefaultTransport),
		},
		tracer: otel.Tracer("notification-service-client"),
	}
}

func (ns *NotificationService) SendOrderCreatedNotification(ctx context.Context, orderID, productID string, quantity int) error {
	ctx, span := ns.tracer.Start(ctx, "send_order_notification")
	defer span.End()

	notificationReq := NotificationRequest{
		Message: fmt.Sprintf("Order %s created successfully for product %s with quantity %d", orderID, productID, quantity),
		OrderID: orderID,
	}

	jsonData, err := json.Marshal(notificationReq)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to marshal notification request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", ns.baseURL+"/api/notifications", bytes.NewBuffer(jsonData))
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Inject trace context into HTTP headers
	otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(req.Header))

	span.SetAttributes(
		attribute.String("notification.order_id", orderID),
		attribute.String("notification.product_id", productID),
		attribute.Int("notification.quantity", quantity),
		attribute.String("notification.url", ns.baseURL+"/api/notifications"),
	)

	resp, err := ns.httpClient.Do(req)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to send notification request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		err := fmt.Errorf("notification service returned status: %d", resp.StatusCode)
		span.RecordError(err)
		return err
	}

	span.SetAttributes(
		attribute.Int("http.status_code", resp.StatusCode),
		attribute.String("http.method", "POST"),
	)

	return nil
}

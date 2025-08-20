# OpenTelemetry Microservices Demo

A complete microservices observability stack with Go services, OpenTelemetry instrumentation, and full monitoring capabilities including distributed tracing, metrics, structured logging, and visualization dashboards.

## 🚀 Features

- **Microservices**: Order Service & Notification Service with HTTP APIs
- **Distributed Tracing**: Full request flow visibility across services
- **Structured Logging**: JSON logs with trace correlation in Loki
- **Metrics & Monitoring**: Custom business metrics in Prometheus
- **Database Observability**: Query performance tracking with PostgreSQL
- **Dashboards**: Pre-configured Grafana visualizations
- **API Testing**: Bruno collection with automated test scripts

## 🏗️ Architecture

```
┌─────────────────┐    ┌──────────────────┐
│   Order Service │────│Notification Service│
│    (Port 8080)  │    │    (Port 8081)     │
└─────────────────┘    └──────────────────┘
         │                       │
         └───────────┬───────────┘
                     │
         ┌───────────▼────────────┐
         │   OpenTelemetry        │
         │     Collector          │
         └─┬─────────┬─────────┬──┘
           │         │         │
    ┌──────▼──┐ ┌────▼────┐ ┌──▼─────┐
    │ Jaeger  │ │Prometheus│ │  Loki  │
    │(Traces) │ │(Metrics) │ │ (Logs) │
    └─────────┘ └─────────┘ └────────┘
                     │
              ┌──────▼──────┐
              │   Grafana   │
              │ (Dashboard) │
              └─────────────┘
```

## 🛠️ Quick Start

### Prerequisites
- Docker & Docker Compose
- Bruno (for API testing)

### 1. Launch Services
```bash
docker compose up -d
```

### 2. Access Dashboards
- **Grafana**: http://localhost:3000 (admin/admin)
- **Jaeger**: http://localhost:16686
- **Prometheus**: http://localhost:9090

### 3. Test APIs

#### Create Product
```bash
curl -X POST http://localhost:8080/products \
  -H "Content-Type: application/json" \
  -d '{"name": "Gaming Laptop", "price": 1299.99, "stock": 15}'
```

#### Create Order
```bash
curl -X POST http://localhost:8080/orders \
  -H "Content-Type: application/json" \
  -d '{"product_id": "PRODUCT_ID", "quantity": 2}'
```

#### Get All Orders (with pagination)
```bash
curl "http://localhost:8080/orders?page=1&page_size=10"
```

## 📊 Observability Features

### Distributed Tracing
- **End-to-end visibility**: Track requests across all microservices
- **Database operations**: Automatic query performance monitoring
- **HTTP calls**: Client-side instrumentation for external API calls
- **Error tracking**: Automatic error correlation with traces

### Structured Logging
- **JSON format**: Machine-readable logs with structured fields
- **Trace correlation**: Every log includes `trace_id` and `span_id`
- **Business context**: Logs include `order_id`, `product_id`, `request_id`
- **Performance data**: Request latencies, database timing, HTTP status codes

### Metrics & Alerts
- **Business metrics**: Order counts, product inventory, notification success rates
- **System metrics**: Database connections, HTTP request rates, error percentages
- **Custom dashboards**: Pre-built Grafana panels for key KPIs

### Database Monitoring
- **Query performance**: Automatic timing for all database operations
- **Connection pooling**: Active connection monitoring
- **Error tracking**: Failed queries with full context

## 🧪 API Testing with Bruno

The repository includes a complete Bruno collection with automated test scripts:

### Features
- **Variable management**: Automatic extraction of IDs between requests
- **Complete workflow**: Create product → Create order → Get order details
- **Validation tests**: Response structure and data validation
- **Trace correlation**: Request IDs for debugging

### Running Tests
1. Open Bruno and import the collection from `./bruno/`
2. Run requests in sequence to see the full workflow
3. Check Grafana/Jaeger for observability data

## 🔍 Monitoring Queries

### Grafana Log Queries
```logql
# All errors across services
{level="error"}

# Trace-specific logs
{trace_id="YOUR_TRACE_ID"}

# Database operations
{service="order-service"} |= "database"

# Order workflow
{order_id="YOUR_ORDER_ID"}
```

### Prometheus Metrics
```promql
# Request rate
rate(http_requests_total[5m])

# Error rate
rate(http_requests_total{status=~"5.."}[5m])

# Database query duration
histogram_quantile(0.95, rate(db_query_duration_seconds_bucket[5m]))
```

## 🗄️ Data Persistence

All data persists across container restarts:
- **Grafana**: Dashboards, users, settings
- **Prometheus**: Metrics history
- **PostgreSQL**: Application data
- **Loki**: Log history

## 🔧 Configuration

### Environment Variables
- `LOG_FORMAT`: Set to "json" for structured logging
- `ENVIRONMENT`: Set to "production" for optimal observability
- `OTEL_EXPORTER_OTLP_ENDPOINT`: OpenTelemetry collector endpoint

### Service Endpoints
- Order Service: http://localhost:8080
- Notification Service: http://localhost:8081
- PostgreSQL: localhost:5432

## 🚦 Health Checks

Each service exposes health endpoints:
- `/health` - Service health status
- `/ready` - Readiness probe for Kubernetes

## 📈 Production Considerations

This demo includes production-ready patterns:
- **Graceful shutdown** with proper cleanup
- **Connection pooling** for database efficiency
- **Rate limiting** and error handling
- **Security headers** and request validation
- **Structured configuration** with environment variables

## 🔍 Troubleshooting

### No logs in Grafana?
1. Check Promtail logs: `docker logs promtail`
2. Verify LOG_FORMAT=json in environment
3. Restart Promtail: `docker restart promtail`

### Missing traces in Jaeger?
1. Verify OTEL_EXPORTER_OTLP_ENDPOINT is set correctly
2. Check OpenTelemetry collector logs
3. Ensure services are making HTTP requests

### Database connection issues?
1. Check PostgreSQL logs: `docker logs postgres`
2. Verify DB_DSN connection string
3. Check database health in service logs

---

This setup demonstrates enterprise-grade observability practices suitable for production microservices environments.
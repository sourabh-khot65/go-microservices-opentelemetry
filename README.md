# Go Microservices OpenTelemetry Demo

This repository demonstrates a production-like Go microservices setup with OpenTelemetry tracing, following best practices for project structure and observability.

## Structure

```
order-service/
  cmd/server/main.go         # Entrypoint
  internal/
    api/                    # HTTP handlers
    service/                # Business logic
    repository/             # Persistence
    models/                 # Data models
    otel/                   # OTel setup
    config/                 # Config loader
  pkg/httpclient/           # Reusable HTTP client
  go.mod
notification-service/
  ... (same structure as order-service)
docker-compose.yaml         # Runs both services, OTel Collector, Jaeger
otel-collector-config.yaml  # OTel Collector config
```

## Running the Demo

1. Build and start all services:
   ```sh
   docker-compose up --build
   ```
2. Access Jaeger UI at [http://localhost:16686](http://localhost:16686)
3. Call the services:
   - Order: `curl localhost:8080/orders`
   - Notification: `curl localhost:8081/notifications`

## References
- [Go Project Layout](https://github.com/golang-standards/project-layout)
- [OpenTelemetry Go](https://github.com/open-telemetry/opentelemetry-go)
- [Jaeger](https://www.jaegertracing.io/)

---

This structure is designed for clarity, maintainability, and easy demonstration of observability best practices in Go microservices.

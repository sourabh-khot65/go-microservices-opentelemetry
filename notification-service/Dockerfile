# syntax=docker/dockerfile:1

FROM golang:1.23-alpine AS builder

WORKDIR /app
COPY go.mod ./
COPY go.sum ./
RUN go mod tidy

COPY . ./
RUN go build -o server ./cmd/server

FROM alpine:3.19
WORKDIR /app
COPY --from=builder /app/server .
COPY --from=builder /app/internal /app/internal
COPY --from=builder /app/pkg /app/pkg
COPY --from=builder /app/go.mod .
COPY --from=builder /app/go.sum .
EXPOSE 8081
CMD ["./server"]

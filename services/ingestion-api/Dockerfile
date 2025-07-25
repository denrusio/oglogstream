# syntax=docker/dockerfile:1
FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY pkg /pkg
COPY services/ingestion-api/go.mod services/ingestion-api/go.sum ./
RUN go mod download
COPY services/ingestion-api/ .
RUN go build -o ingestion-api main.go

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/ingestion-api .
EXPOSE 8080
CMD ["./ingestion-api"] 
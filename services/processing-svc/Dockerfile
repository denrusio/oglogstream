# syntax=docker/dockerfile:1
FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY pkg /pkg
COPY services/processing-svc/go.mod services/processing-svc/go.sum ./
RUN go mod download
COPY services/processing-svc/ .
RUN go build -o processing-svc main.go

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/processing-svc .
CMD ["./processing-svc"] 
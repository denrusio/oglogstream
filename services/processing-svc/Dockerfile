# syntax=docker/dockerfile:1
FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY pkg /pkg
COPY services/processing-svc/go.mod services/processing-svc/go.sum ./
RUN go mod download
COPY services/processing-svc/ .
RUN go build -o processing-svc main.go

FROM alpine:latest
RUN apk add --no-cache curl
WORKDIR /root/
COPY --from=builder /app/processing-svc .
EXPOSE 8082
CMD ["./processing-svc"] 
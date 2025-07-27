# syntax=docker/dockerfile:1
FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY pkg /pkg
COPY services/query-api/go.mod services/query-api/go.sum ./
RUN go mod download
COPY services/query-api/ .
RUN go build -o query-api main.go

FROM alpine:latest
RUN apk add --no-cache curl
WORKDIR /root/
COPY --from=builder /app/query-api .
EXPOSE 8081
CMD ["./query-api"] 
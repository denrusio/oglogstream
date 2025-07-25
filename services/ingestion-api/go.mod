module github.com/yourusername/oglogstream-ingestion-api

go 1.24.5

require (
	github.com/go-chi/chi/v5 v5.2.2
	github.com/nats-io/nats.go v1.43.0
	github.com/yourusername/oglogstream-models v0.0.0-00010101000000-000000000000
)

require (
	github.com/klauspost/compress v1.18.0 // indirect
	github.com/nats-io/nkeys v0.4.11 // indirect
	github.com/nats-io/nuid v1.0.1 // indirect
	golang.org/x/crypto v0.40.0 // indirect
	golang.org/x/sys v0.34.0 // indirect
)

replace github.com/yourusername/oglogstream-ingestion-api/pkg/models => ../../pkg/models

replace github.com/yourusername/oglogstream-models => ../../pkg/models

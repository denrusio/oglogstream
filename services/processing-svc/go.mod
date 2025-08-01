module github.com/yourusername/oglogstream-processing-svc

go 1.24.5

require (
	github.com/ClickHouse/clickhouse-go/v2 v2.39.0
	github.com/go-chi/chi/v5 v5.0.12
	github.com/nats-io/nats.go v1.43.0
	github.com/yourusername/oglogstream-models v0.0.0
)

require (
	github.com/ClickHouse/ch-go v0.67.0 // indirect
	github.com/andybalholm/brotli v1.2.0 // indirect
	github.com/go-faster/city v1.0.1 // indirect
	github.com/go-faster/errors v0.7.1 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/klauspost/compress v1.18.0 // indirect
	github.com/nats-io/nkeys v0.4.11 // indirect
	github.com/nats-io/nuid v1.0.1 // indirect
	github.com/paulmach/orb v0.11.1 // indirect
	github.com/pierrec/lz4/v4 v4.1.22 // indirect
	github.com/segmentio/asm v1.2.0 // indirect
	github.com/shopspring/decimal v1.4.0 // indirect
	go.opentelemetry.io/otel v1.37.0 // indirect
	go.opentelemetry.io/otel/trace v1.37.0 // indirect
	golang.org/x/crypto v0.40.0 // indirect
	golang.org/x/sys v0.34.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/yourusername/oglogstream-models => ../../pkg/models

replace github.com/yourusername/oglogstream-processing-svc/pkg/models => ../../pkg/models

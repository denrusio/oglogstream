package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/nats-io/nats.go"

	"github.com/yourusername/oglogstream-models"
)

const (
	maxMessageSize = 10 * 1024     // 10KB max message
	maxServiceSize = 100           // 100 chars max service name  
	shutdownTimeout = 30 * time.Second
)

var validLevels = map[string]bool{
	"debug": true,
	"info":  true,
	"warn":  true,
	"error": true,
	"fatal": true,
}

// CORS middleware
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		
		next.ServeHTTP(w, r)
	})
}

// Request size limiting middleware
func maxBytesMiddleware(maxBytes int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
			next.ServeHTTP(w, r)
		})
	}
}

func validateLogEntry(entry *models.LogEntry) error {
	// Validate level
	if entry.Level == "" {
		return fmt.Errorf("level is required")
	}
	
	level := strings.ToLower(entry.Level)
	if !validLevels[level] {
		return fmt.Errorf("invalid level '%s', must be one of: debug, info, warn, error, fatal", entry.Level)
	}
	entry.Level = level // normalize to lowercase
	
	// Validate message
	if entry.Message == "" {
		return fmt.Errorf("message is required")
	}
	if len(entry.Message) > maxMessageSize {
		return fmt.Errorf("message too long (max %d characters)", maxMessageSize)
	}
	
	// Validate service
	if entry.Service == "" {
		return fmt.Errorf("service is required")
	}
	if len(entry.Service) > maxServiceSize {
		return fmt.Errorf("service name too long (max %d characters)", maxServiceSize)
	}
	
	// Set timestamp if not provided or invalid
	if entry.Timestamp.IsZero() {
		entry.Timestamp = time.Now().UTC()
	}
	
	return nil
}

func createLogHandler(nc *nats.Conn) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var entry models.LogEntry
		
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields() // Strict JSON parsing
		
		if err := decoder.Decode(&entry); err != nil {
			http.Error(w, fmt.Sprintf("Invalid JSON: %v", err), http.StatusBadRequest)
			return
		}
		
		// Validate the log entry
		if err := validateLogEntry(&entry); err != nil {
			http.Error(w, fmt.Sprintf("Validation error: %v", err), http.StatusBadRequest)
			return
		}
		
		// Marshal to JSON
		data, err := json.Marshal(entry)
		if err != nil {
			log.Printf("Marshal error: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		
		// Publish to NATS
		if err := nc.Publish("logs.raw", data); err != nil {
			log.Printf("NATS publish error: %v", err)
			http.Error(w, "Message delivery failed", http.StatusServiceUnavailable)
			return
		}
		
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		w.Write([]byte(`{"status":"accepted","timestamp":"` + entry.Timestamp.Format(time.RFC3339) + `"}`))
	}
}

func main() {
	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		natsURL = "nats://localhost:4222"
	}

	// Connect to NATS with retries and better options
	opts := []nats.Option{
		nats.ReconnectWait(2 * time.Second),
		nats.MaxReconnects(5),
		nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
			log.Printf("NATS disconnected: %v", err)
		}),
		nats.ReconnectHandler(func(nc *nats.Conn) {
			log.Printf("NATS reconnected to %v", nc.ConnectedUrl())
		}),
	}
	
	nc, err := nats.Connect(natsURL, opts...)
	if err != nil {
		log.Fatalf("Failed to connect to NATS: %v", err)
	}
	defer nc.Drain()

	// Setup router
	r := chi.NewRouter()
	
	// Middleware stack
	r.Use(corsMiddleware)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(maxBytesMiddleware(50*1024)) // 50KB max request size
	r.Use(middleware.Timeout(30 * time.Second))

	// Health check endpoint
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		status := "ok"
		if nc.Status() != nats.CONNECTED {
			status = "degraded"
			w.WriteHeader(http.StatusServiceUnavailable)
		}
		
		w.Header().Set("Content-Type", "application/json")
		response := fmt.Sprintf(`{"status":"%s","service":"ingestion-api","nats_status":"%s","timestamp":"%s"}`, 
			status, nc.Status(), time.Now().UTC().Format(time.RFC3339))
		w.Write([]byte(response))
	})

	// Log ingestion endpoint
	r.Post("/log", createLogHandler(nc))

	// Setup HTTP server
	addr := ":8080"
	srv := &http.Server{
		Addr:           addr,
		Handler:        r,
		ReadTimeout:    15 * time.Second,
		WriteTimeout:   15 * time.Second,
		IdleTimeout:    60 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1MB
	}

	// Start server in goroutine
	go func() {
		log.Printf("Ingestion API listening on %s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Setup graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Wait for shutdown signal
	<-sigChan
	log.Println("Shutdown signal received, stopping gracefully...")

	// Create shutdown context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	// Shutdown HTTP server
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}

	log.Println("Ingestion API stopped")
} 
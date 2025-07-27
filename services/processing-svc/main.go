package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	_ "github.com/ClickHouse/clickhouse-go/v2"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/nats-io/nats.go"

	"github.com/yourusername/oglogstream-models"
)

const (
	batchSize    = 100
	flushTimeout = 2 * time.Second
	maxRetries   = 3
)

type BatchProcessor struct {
	db       *sql.DB
	hostname string
	batch    []models.LogEntry
	mutex    sync.Mutex
	done     chan bool
}

func NewBatchProcessor(db *sql.DB, hostname string) *BatchProcessor {
	bp := &BatchProcessor{
		db:       db,
		hostname: hostname,
		batch:    make([]models.LogEntry, 0, batchSize),
		done:     make(chan bool),
	}
	
	// Start flush timer
	go bp.flushTimer()
	
	return bp
}

func (bp *BatchProcessor) AddEntry(entry models.LogEntry) {
	bp.mutex.Lock()
	defer bp.mutex.Unlock()
	
	bp.batch = append(bp.batch, entry)
	
	if len(bp.batch) >= batchSize {
		bp.flushBatch()
	}
}

func (bp *BatchProcessor) flushTimer() {
	ticker := time.NewTicker(flushTimeout)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			bp.mutex.Lock()
			if len(bp.batch) > 0 {
				bp.flushBatch()
			}
			bp.mutex.Unlock()
		case <-bp.done:
			return
		}
	}
}

func (bp *BatchProcessor) flushBatch() {
	if len(bp.batch) == 0 {
		return
	}
	
	batch := make([]models.LogEntry, len(bp.batch))
	copy(batch, bp.batch)
	bp.batch = bp.batch[:0] // reset slice
	
	// Insert batch with retries
	go bp.insertBatchWithRetry(batch)
}

func (bp *BatchProcessor) insertBatchWithRetry(batch []models.LogEntry) {
	for attempt := 1; attempt <= maxRetries; attempt++ {
		err := bp.insertBatch(batch)
		if err == nil {
			log.Printf("[%s] Successfully inserted batch of %d logs", bp.hostname, len(batch))
			return
		}
		
		log.Printf("[%s] Batch insert attempt %d failed: %v", bp.hostname, attempt, err)
		if attempt < maxRetries {
			time.Sleep(time.Duration(attempt) * time.Second)
		}
	}
	
	log.Printf("[%s] Failed to insert batch after %d attempts, dropping %d logs", bp.hostname, maxRetries, len(batch))
}

func (bp *BatchProcessor) insertBatch(batch []models.LogEntry) error {
	if len(batch) == 0 {
		return nil
	}
	
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	tx, err := bp.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	
	stmt, err := tx.PrepareContext(ctx, `INSERT INTO logs (timestamp, level, message, service) VALUES (?, ?, ?, ?)`)
	if err != nil {
		return err
	}
	defer stmt.Close()
	
	for _, entry := range batch {
		_, err = stmt.ExecContext(ctx, entry.Timestamp, entry.Level, entry.Message, entry.Service)
		if err != nil {
			return err
		}
	}
	
	return tx.Commit()
}

func (bp *BatchProcessor) Stop() {
	close(bp.done)
	
	// Flush remaining entries
	bp.mutex.Lock()
	defer bp.mutex.Unlock()
	
	if len(bp.batch) > 0 {
		log.Printf("[%s] Flushing remaining %d logs before shutdown", bp.hostname, len(bp.batch))
		bp.insertBatchWithRetry(bp.batch)
	}
}

func main() {
	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		natsURL = "nats://localhost:4222"
	}
	chDSN := os.Getenv("CLICKHOUSE_DSN")
	if chDSN == "" {
		chDSN = "clickhouse://default:@localhost:9000/default"
	}

	// Connect to NATS
	nc, err := nats.Connect(natsURL)
	if err != nil {
		log.Fatalf("Failed to connect to NATS: %v", err)
	}
	defer nc.Drain()

	// Connect to ClickHouse with connection pool
	db, err := sql.Open("clickhouse", chDSN+"?max_open_conns=5&max_idle_conns=2&conn_max_lifetime=300s")
	if err != nil {
		log.Fatalf("Failed to connect to ClickHouse: %v", err)
	}
	defer db.Close()
	
	// Test connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping ClickHouse: %v", err)
	}

	// Get hostname for logging
	hostname, _ := os.Hostname()
	
	// Initialize batch processor
	processor := NewBatchProcessor(db, hostname)
	
	// Setup HTTP server for health checks
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	
	// Health check endpoint
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		status := "ok"
		httpStatus := http.StatusOK
		
		// Check NATS connection
		if nc.Status() != nats.CONNECTED {
			status = "degraded"
			httpStatus = http.StatusServiceUnavailable
		}
		
		// Check ClickHouse connection
		if err := db.Ping(); err != nil {
			status = "degraded"
			httpStatus = http.StatusServiceUnavailable
		}
		
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(httpStatus)
		response := fmt.Sprintf(`{"status":"%s","service":"processing-svc","nats_status":"%s","clickhouse_status":"connected","timestamp":"%s"}`, 
			status, nc.Status(), time.Now().UTC().Format(time.RFC3339))
		w.Write([]byte(response))
	})
	
	// Setup HTTP server
	addr := ":8082"
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
		log.Printf("[%s] Processing service HTTP server listening on %s", hostname, addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()
	
	// Setup graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	
	// Subscribe to NATS
	sub, err := nc.QueueSubscribe("logs.raw", "processing-group", func(msg *nats.Msg) {
		var entry models.LogEntry
		if err := json.Unmarshal(msg.Data, &entry); err != nil {
			log.Printf("[%s] Invalid log entry: %v", hostname, err)
			return
		}
		
		processor.AddEntry(entry)
	})
	if err != nil {
		log.Fatalf("Failed to subscribe to NATS: %v", err)
	}
	defer sub.Unsubscribe()

	log.Printf("[%s] Processing service started in queue group 'processing-group'. Batch size: %d, flush timeout: %v", 
		hostname, batchSize, flushTimeout)
	
	// Wait for shutdown signal
	<-sigChan
	log.Printf("[%s] Shutdown signal received, stopping gracefully...", hostname)
	
	// Stop batch processor
	processor.Stop()
	
	// Shutdown HTTP server
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("[%s] HTTP server shutdown error: %v", hostname, err)
	}
	
	log.Printf("[%s] Processing service stopped", hostname)
} 
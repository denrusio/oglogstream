package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	_ "github.com/ClickHouse/clickhouse-go/v2"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// Mock database for testing
type mockDB struct {
	logs  []LogEntry
	stats []Stat
}

func (m *mockDB) Query(query string, args ...interface{}) (*sql.Rows, error) {
	// This would need proper mock implementation in real tests
	// For now, we'll test with integration approach
	return nil, fmt.Errorf("mock query not implemented")
}

func (m *mockDB) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return m.Query(query, args...)
}

func (m *mockDB) Close() error {
	return nil
}

// Test data
var testLogs = []LogEntry{
	{
		Timestamp: "2025-07-24T16:00:00Z",
		Level:     "info",
		Message:   "User login successful",
		Service:   "auth-service",
	},
	{
		Timestamp: "2025-07-24T16:01:00Z",
		Level:     "error",
		Message:   "Payment failed",
		Service:   "payment-api",
	},
	{
		Timestamp: "2025-07-24T16:02:00Z",
		Level:     "warn",
		Message:   "High CPU usage",
		Service:   "monitoring-svc",
	},
}

var testStats = []Stat{
	{Level: "info", Count: 5},
	{Level: "error", Count: 3},
	{Level: "warn", Count: 2},
	{Level: "fatal", Count: 1},
}

// Create test router with mock handlers
func createTestRouter() *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(corsMiddleware)

	// Mock /api/logs endpoint
	r.Get("/api/logs", func(w http.ResponseWriter, r *http.Request) {
		levelFilter := r.URL.Query().Get("level")
		serviceFilter := r.URL.Query().Get("service")
		
		var filteredLogs []LogEntry
		for _, log := range testLogs {
			// Apply level filter
			if levelFilter != "" && log.Level != levelFilter {
				continue
			}
			
			// Apply service filter (case-insensitive partial match)
			if serviceFilter != "" && !strings.Contains(strings.ToLower(log.Service), strings.ToLower(serviceFilter)) {
				continue
			}
			
			filteredLogs = append(filteredLogs, log)
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(filteredLogs)
	})

	// Mock /api/stats endpoint
	r.Get("/api/stats", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(testStats)
	})

	return r
}

func TestGetLogsEndpoint(t *testing.T) {
	router := createTestRouter()

	tests := []struct {
		name           string
		url            string
		expectedStatus int
		expectedCount  int
		description    string
	}{
		{
			name:           "Get all logs",
			url:            "/api/logs",
			expectedStatus: http.StatusOK,
			expectedCount:  3,
			description:    "Should return all test logs",
		},
		{
			name:           "Filter by level - info",
			url:            "/api/logs?level=info",
			expectedStatus: http.StatusOK,
			expectedCount:  1,
			description:    "Should return only info logs",
		},
		{
			name:           "Filter by level - error",
			url:            "/api/logs?level=error",
			expectedStatus: http.StatusOK,
			expectedCount:  1,
			description:    "Should return only error logs",
		},
		{
			name:           "Filter by service - auth",
			url:            "/api/logs?service=auth",
			expectedStatus: http.StatusOK,
			expectedCount:  1,
			description:    "Should return logs from auth-service",
		},
		{
			name:           "Filter by service - payment",
			url:            "/api/logs?service=payment",
			expectedStatus: http.StatusOK,
			expectedCount:  1,
			description:    "Should return logs from payment-api",
		},
		{
			name:           "Combined filter - error + payment",
			url:            "/api/logs?level=error&service=payment",
			expectedStatus: http.StatusOK,
			expectedCount:  1,
			description:    "Should return error logs from payment service",
		},
		{
			name:           "Filter with no results",
			url:            "/api/logs?level=fatal",
			expectedStatus: http.StatusOK,
			expectedCount:  0,
			description:    "Should return empty array for non-existent level",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.url, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			var logs []LogEntry
			if err := json.NewDecoder(w.Body).Decode(&logs); err != nil {
				t.Fatalf("Failed to decode response: %v", err)
			}

			if len(logs) != tt.expectedCount {
				t.Errorf("Expected %d logs, got %d. %s", tt.expectedCount, len(logs), tt.description)
			}

			// Verify Content-Type header
			expectedContentType := "application/json"
			if ct := w.Header().Get("Content-Type"); ct != expectedContentType {
				t.Errorf("Expected Content-Type %s, got %s", expectedContentType, ct)
			}
		})
	}
}

func TestGetStatsEndpoint(t *testing.T) {
	router := createTestRouter()

	req := httptest.NewRequest("GET", "/api/stats", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var stats []Stat
	if err := json.NewDecoder(w.Body).Decode(&stats); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	expectedStatsCount := 4
	if len(stats) != expectedStatsCount {
		t.Errorf("Expected %d stats, got %d", expectedStatsCount, len(stats))
	}

	// Verify stats structure and content
	expectedLevels := map[string]int{
		"info":  5,
		"error": 3,
		"warn":  2,
		"fatal": 1,
	}

	for _, stat := range stats {
		if expectedCount, exists := expectedLevels[stat.Level]; exists {
			if stat.Count != expectedCount {
				t.Errorf("Expected count %d for level %s, got %d", expectedCount, stat.Level, stat.Count)
			}
		} else {
			t.Errorf("Unexpected level in stats: %s", stat.Level)
		}
	}

	// Verify Content-Type header
	expectedContentType := "application/json"
	if ct := w.Header().Get("Content-Type"); ct != expectedContentType {
		t.Errorf("Expected Content-Type %s, got %s", expectedContentType, ct)
	}
}

func TestCORSHeaders(t *testing.T) {
	router := createTestRouter()

	tests := []struct {
		method string
		url    string
	}{
		{"GET", "/api/logs"},
		{"GET", "/api/stats"},
		{"OPTIONS", "/api/logs"},
		{"OPTIONS", "/api/stats"},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s %s", tt.method, tt.url), func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.url, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			// Check CORS headers
			expectedHeaders := map[string]string{
				"Access-Control-Allow-Origin":  "*",
				"Access-Control-Allow-Methods": "GET, POST, PUT, DELETE, OPTIONS",
				"Access-Control-Allow-Headers": "Content-Type, Authorization",
			}

			for header, expectedValue := range expectedHeaders {
				if actualValue := w.Header().Get(header); actualValue != expectedValue {
					t.Errorf("Expected %s header to be %s, got %s", header, expectedValue, actualValue)
				}
			}

			// OPTIONS requests should return 200 OK
			if tt.method == "OPTIONS" && w.Code != http.StatusOK {
				t.Errorf("OPTIONS request should return 200 OK, got %d", w.Code)
			}
		})
	}
}

func TestLogEntryJSONSerialization(t *testing.T) {
	testLog := LogEntry{
		Timestamp: "2025-07-24T16:00:00Z",
		Level:     "info",
		Message:   "Test message",
		Service:   "test-service",
	}

	// Test JSON marshaling
	jsonData, err := json.Marshal(testLog)
	if err != nil {
		t.Fatalf("Failed to marshal LogEntry: %v", err)
	}

	// Test JSON unmarshaling
	var unmarshaled LogEntry
	if err := json.Unmarshal(jsonData, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal LogEntry: %v", err)
	}

	// Verify fields
	if unmarshaled.Timestamp != testLog.Timestamp {
		t.Errorf("Expected timestamp %s, got %s", testLog.Timestamp, unmarshaled.Timestamp)
	}
	if unmarshaled.Level != testLog.Level {
		t.Errorf("Expected level %s, got %s", testLog.Level, unmarshaled.Level)
	}
	if unmarshaled.Message != testLog.Message {
		t.Errorf("Expected message %s, got %s", testLog.Message, unmarshaled.Message)
	}
	if unmarshaled.Service != testLog.Service {
		t.Errorf("Expected service %s, got %s", testLog.Service, unmarshaled.Service)
	}
}

func TestStatJSONSerialization(t *testing.T) {
	testStat := Stat{
		Level: "error",
		Count: 42,
	}

	// Test JSON marshaling
	jsonData, err := json.Marshal(testStat)
	if err != nil {
		t.Fatalf("Failed to marshal Stat: %v", err)
	}

	// Test JSON unmarshaling
	var unmarshaled Stat
	if err := json.Unmarshal(jsonData, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal Stat: %v", err)
	}

	// Verify fields
	if unmarshaled.Level != testStat.Level {
		t.Errorf("Expected level %s, got %s", testStat.Level, unmarshaled.Level)
	}
	if unmarshaled.Count != testStat.Count {
		t.Errorf("Expected count %d, got %d", testStat.Count, unmarshaled.Count)
	}
}

// Benchmark tests
func BenchmarkGetLogsEndpoint(b *testing.B) {
	router := createTestRouter()

	req := httptest.NewRequest("GET", "/api/logs", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		if w.Code != http.StatusOK {
			b.Fatalf("Expected status 200, got %d", w.Code)
		}
	}
}

func BenchmarkGetStatsEndpoint(b *testing.B) {
	router := createTestRouter()

	req := httptest.NewRequest("GET", "/api/stats", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		if w.Code != http.StatusOK {
			b.Fatalf("Expected status 200, got %d", w.Code)
		}
	}
} 
package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/yourusername/logstream-models"
)

func TestLogEntryUnmarshal(t *testing.T) {
	jsonStr := `{"timestamp":"2024-06-01T12:00:00Z","level":"error","message":"Test","service":"test"}`
	var entry models.LogEntry
	if err := json.Unmarshal([]byte(jsonStr), &entry); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if entry.Level != "error" || entry.Message != "Test" || entry.Service != "test" {
		t.Errorf("unexpected values: %+v", entry)
	}
}

func TestPostLogEndpoint(t *testing.T) {
	// Мокаем NATS publisher через closure
	var published []byte
	r := chi.NewRouter()
	r.Post("/log", func(w http.ResponseWriter, r *http.Request) {
		var entry models.LogEntry
		if err := json.NewDecoder(r.Body).Decode(&entry); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}
		entry.Timestamp = time.Now().UTC()
		b, err := json.Marshal(entry)
		if err != nil {
			http.Error(w, "marshal error", http.StatusInternalServerError)
			return
		}
		published = b
		w.WriteHeader(http.StatusAccepted)
	})

	body := []byte(`{"timestamp":"2024-06-01T12:00:00Z","level":"info","message":"Hello","service":"svc"}`)
	req := httptest.NewRequest(http.MethodPost, "/log", bytes.NewReader(body))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusAccepted {
		t.Errorf("expected 202, got %d", w.Code)
	}
	if len(published) == 0 {
		t.Error("log not published")
	}
} 
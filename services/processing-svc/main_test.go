package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"testing"

	_ "github.com/ClickHouse/clickhouse-go/v2"
	"github.com/yourusername/oglogstream-models"
)

type mockDB struct {
	lastQuery string
	lastArgs  []interface{}
}

func (m *mockDB) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	m.lastQuery = query
	m.lastArgs = args
	return nil, nil
}

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

func TestInsertLogToClickHouse(t *testing.T) {
	mdb := &mockDB{}
	entry := models.LogEntry{
		Level:   "info",
		Message: "msg",
		Service: "svc",
	}
	_, err := mdb.ExecContext(context.Background(),
		`INSERT INTO logs (timestamp, level, message, service) VALUES (?, ?, ?, ?)`,
		entry.Timestamp, entry.Level, entry.Message, entry.Service,
	)
	if err != nil {
		t.Fatalf("insert failed: %v", err)
	}
	if mdb.lastQuery == "" || len(mdb.lastArgs) != 4 {
		t.Errorf("unexpected query or args: %v %v", mdb.lastQuery, mdb.lastArgs)
	}
} 
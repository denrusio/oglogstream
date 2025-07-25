package models

import "time"

// LogEntry описывает структуру лога для всех сервисов OgLogStream
// Используется для передачи, обработки и хранения логов

type LogEntry struct {
	Timestamp time.Time `json:"timestamp"`
	Level     string    `json:"level"`    // e.g., "info", "error"
	Message   string    `json:"message"`
	Service   string    `json:"service"`  // e.g., "auth-service"
} 
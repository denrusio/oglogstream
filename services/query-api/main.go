package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	_ "github.com/ClickHouse/clickhouse-go/v2"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/gorilla/websocket"
	"github.com/nats-io/nats.go"
)

type LogEntry struct {
	Timestamp string `json:"timestamp"`
	Level     string `json:"level"`
	Message   string `json:"message"`
	Service   string `json:"service"`
}

type Stat struct {
	Level string `json:"level"`
	Count int    `json:"count"`
}

// Client represents a single WebSocket client
type Client struct {
	hub  *Hub
	conn *websocket.Conn
	send chan []byte
}

// Hub maintains the set of active clients and broadcasts messages to them
type Hub struct {
	// Registered clients
	clients map[*Client]bool

	// Inbound messages from clients
	broadcast chan []byte

	// Register requests from clients
	register chan *Client

	// Unregister requests from clients
	unregister chan *Client
}

func newHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
			log.Printf("Client connected. Total: %d", len(h.clients))

		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
				log.Printf("Client disconnected. Total: %d", len(h.clients))
			}

		case message := <-h.broadcast:
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					// Client's send channel is full, remove it
					delete(h.clients, client)
					close(client.send)
				}
			}
		}
	}
}

// readPump pumps messages from the websocket connection to the hub
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(512)
	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, _, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}
	}
}

// writePump pumps messages from the hub to the websocket connection
func (c *Client) writePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				// The hub closed the channel
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
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

func main() {
	chDSN := os.Getenv("CLICKHOUSE_DSN")
	if chDSN == "" {
		chDSN = "clickhouse://default:@clickhouse:9000/default"
	}
	db, err := sql.Open("clickhouse", chDSN)
	if err != nil {
		log.Fatalf("Failed to connect to ClickHouse: %v", err)
	}
	defer db.Close()

	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		natsURL = "nats://nats:4222"
	}
	nc, err := nats.Connect(natsURL)
	if err != nil {
		log.Fatalf("Failed to connect to NATS: %v", err)
	}
	defer nc.Drain()

	// Create and start the hub
	hub := newHub()
	go hub.run()

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(corsMiddleware)

	// Health check endpoint
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok","service":"query-api"}`))
	})

	r.Get("/api/logs", func(w http.ResponseWriter, r *http.Request) {
		// Get query parameters for filtering
		levelFilter := r.URL.Query().Get("level")
		serviceFilter := r.URL.Query().Get("service")
		
		// Build dynamic SQL query with filters
		query := `SELECT timestamp, level, message, service FROM logs`
		var conditions []string
		var args []interface{}
		
		if levelFilter != "" {
			conditions = append(conditions, "level = ?")
			args = append(args, levelFilter)
		}
		
		if serviceFilter != "" {
			conditions = append(conditions, "service ILIKE ?")
			args = append(args, "%"+serviceFilter+"%")
		}
		
		if len(conditions) > 0 {
			query += " WHERE " + strings.Join(conditions, " AND ")
		}
		
		query += " ORDER BY timestamp DESC LIMIT 100"
		
		rows, err := db.Query(query, args...)
		if err != nil {
			log.Printf("DB error (logs): %v", err)
			http.Error(w, "db error", http.StatusInternalServerError)
			return
		}
		defer rows.Close()
		var logs []LogEntry
		for rows.Next() {
			var e LogEntry
			if err := rows.Scan(&e.Timestamp, &e.Level, &e.Message, &e.Service); err != nil {
				continue
			}
			logs = append(logs, e)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(logs)
	})

	r.Get("/api/stats", func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query(`SELECT level, count() FROM logs GROUP BY level`)
		if err != nil {
			log.Printf("DB error (stats): %v", err)
			http.Error(w, "db error", http.StatusInternalServerError)
			return
		}
		defer rows.Close()
		var stats []Stat
		for rows.Next() {
			var s Stat
			if err := rows.Scan(&s.Level, &s.Count); err != nil {
				continue
			}
			stats = append(stats, s)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(stats)
	})

	r.Get("/ws/live", func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		}
		
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("WebSocket upgrade error: %v", err)
			return
		}

		client := &Client{
			hub:  hub,
			conn: conn,
			send: make(chan []byte, 256),
		}

		client.hub.register <- client

		// Start goroutines for reading and writing
		go client.writePump()
		go client.readPump()
	})

	// Subscribe to NATS and broadcast to all clients
	_, err = nc.Subscribe("logs.raw", func(msg *nats.Msg) {
		hub.broadcast <- msg.Data
	})
	if err != nil {
		log.Fatalf("Failed to subscribe to NATS: %v", err)
	}

	addr := ":8081"
	log.Printf("Query API listening on %s", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("Server error: %v", err)
	}
} 
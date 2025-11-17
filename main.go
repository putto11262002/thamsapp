package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
)

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true // Allow all origins for local network
		},
	}

	// Connected clients map (IP -> WebSocket connection)
	clients   = make(map[string]*websocket.Conn)
	clientsMu sync.RWMutex

	// Broadcast channel for messages
	broadcast = make(chan Message)
)

// MessagePayload represents the JSON structure for WebSocket messages
type MessagePayload struct {
	Type    string  `json:"type"`
	Message Message `json:"message,omitempty"`
	History []Message `json:"history,omitempty"`
	ClientIP string `json:"client_ip,omitempty"`
}

// getClientIP extracts the client's IP address from the request
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header first
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		ips := strings.Split(forwarded, ",")
		return strings.TrimSpace(ips[0])
	}

	// Check X-Real-IP header
	realIP := r.Header.Get("X-Real-IP")
	if realIP != "" {
		return realIP
	}

	// Fall back to RemoteAddr
	ip := r.RemoteAddr
	// Remove port if present
	if idx := strings.LastIndex(ip, ":"); idx != -1 {
		ip = ip[:idx]
	}
	return ip
}

// handleWebSocket handles WebSocket connections
func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	clientIP := getClientIP(r)
	log.Printf("New client connected: %s", clientIP)

	// Save client to database
	if err := SaveClient(clientIP); err != nil {
		log.Printf("Error saving client: %v", err)
	}

	// Register client
	clientsMu.Lock()
	clients[clientIP] = conn
	clientsMu.Unlock()

	// Send connection confirmation and recent message history
	recentMessages, err := GetRecentMessages(50)
	if err != nil {
		log.Printf("Error fetching recent messages: %v", err)
		recentMessages = []Message{}
	}

	welcomePayload := MessagePayload{
		Type:     "connected",
		ClientIP: clientIP,
		History:  recentMessages,
	}

	if err := conn.WriteJSON(welcomePayload); err != nil {
		log.Printf("Error sending welcome message: %v", err)
	}

	// Handle incoming messages
	go handleMessages(conn, clientIP)
}

// handleMessages reads messages from a WebSocket connection
func handleMessages(conn *websocket.Conn, clientIP string) {
	defer func() {
		// Unregister client
		clientsMu.Lock()
		delete(clients, clientIP)
		clientsMu.Unlock()
		conn.Close()
		log.Printf("Client disconnected: %s", clientIP)
	}()

	for {
		var payload map[string]interface{}
		err := conn.ReadJSON(&payload)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		// Extract message content
		content, ok := payload["content"].(string)
		if !ok || content == "" {
			continue
		}

		// Save message to database
		msg, err := SaveMessage(clientIP, content)
		if err != nil {
			log.Printf("Error saving message: %v", err)
			continue
		}

		// Broadcast message to all clients
		broadcast <- *msg
	}
}

// handleBroadcast sends messages to all connected clients
func handleBroadcast() {
	for msg := range broadcast {
		payload := MessagePayload{
			Type:    "message",
			Message: msg,
		}

		clientsMu.RLock()
		for ip, conn := range clients {
			err := conn.WriteJSON(payload)
			if err != nil {
				log.Printf("Error broadcasting to %s: %v", ip, err)
				conn.Close()
				clientsMu.RUnlock()
				clientsMu.Lock()
				delete(clients, ip)
				clientsMu.Unlock()
				clientsMu.RLock()
			}
		}
		clientsMu.RUnlock()
	}
}

// handleIndex serves the main HTML page
func handleIndex(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "index.html")
}

// handleAPI provides a REST API for clients and messages
func handleAPI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.URL.Path {
	case "/api/clients":
		clients, err := GetAllClients()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(clients)

	case "/api/messages":
		messages, err := GetRecentMessages(100)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(messages)

	default:
		http.NotFound(w, r)
	}
}

func main() {
	// Initialize database
	if err := InitDB(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Start broadcast handler
	go handleBroadcast()

	// Setup routes
	http.HandleFunc("/", handleIndex)
	http.HandleFunc("/ws", handleWebSocket)
	http.HandleFunc("/api/", handleAPI)

	// Start server
	port := ":8080"
	log.Printf("Server starting on http://0.0.0.0%s", port)
	log.Printf("WebSocket endpoint: ws://0.0.0.0%s/ws", port)

	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

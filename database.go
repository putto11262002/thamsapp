package main

import (
	"database/sql"
	"log"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Message struct {
	ID        int64     `json:"id"`
	ClientIP  string    `json:"client_ip"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

type Client struct {
	IP         string    `json:"ip"`
	LastSeen   time.Time `json:"last_seen"`
	FirstSeen  time.Time `json:"first_seen"`
}

var db *sql.DB

// InitDB initializes the SQLite database
func InitDB() error {
	var err error
	db, err = sql.Open("sqlite3", "./chat.db")
	if err != nil {
		return err
	}

	// Create clients table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS clients (
			ip TEXT PRIMARY KEY,
			first_seen DATETIME NOT NULL,
			last_seen DATETIME NOT NULL
		)
	`)
	if err != nil {
		return err
	}

	// Create messages table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS messages (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			client_ip TEXT NOT NULL,
			content TEXT NOT NULL,
			timestamp DATETIME NOT NULL,
			FOREIGN KEY (client_ip) REFERENCES clients(ip)
		)
	`)
	if err != nil {
		return err
	}

	log.Println("Database initialized successfully")
	return nil
}

// SaveClient saves or updates a client in the database
func SaveClient(ip string) error {
	now := time.Now()
	_, err := db.Exec(`
		INSERT INTO clients (ip, first_seen, last_seen)
		VALUES (?, ?, ?)
		ON CONFLICT(ip) DO UPDATE SET last_seen = ?
	`, ip, now, now, now)
	return err
}

// SaveMessage saves a message to the database
func SaveMessage(clientIP, content string) (*Message, error) {
	now := time.Now()
	result, err := db.Exec(`
		INSERT INTO messages (client_ip, content, timestamp)
		VALUES (?, ?, ?)
	`, clientIP, content, now)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return &Message{
		ID:        id,
		ClientIP:  clientIP,
		Content:   content,
		Timestamp: now,
	}, nil
}

// GetRecentMessages retrieves the last N messages from the database
func GetRecentMessages(limit int) ([]Message, error) {
	rows, err := db.Query(`
		SELECT id, client_ip, content, timestamp
		FROM messages
		ORDER BY timestamp DESC
		LIMIT ?
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []Message
	for rows.Next() {
		var msg Message
		err := rows.Scan(&msg.ID, &msg.ClientIP, &msg.Content, &msg.Timestamp)
		if err != nil {
			return nil, err
		}
		messages = append(messages, msg)
	}

	// Reverse to get chronological order
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	return messages, nil
}

// GetAllClients retrieves all clients from the database
func GetAllClients() ([]Client, error) {
	rows, err := db.Query(`
		SELECT ip, first_seen, last_seen
		FROM clients
		ORDER BY last_seen DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var clients []Client
	for rows.Next() {
		var client Client
		err := rows.Scan(&client.IP, &client.FirstSeen, &client.LastSeen)
		if err != nil {
			return nil, err
		}
		clients = append(clients, client)
	}

	return clients, nil
}

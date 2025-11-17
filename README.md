# Local Network Chat Application

A real-time messaging application for local WiFi networks built with Go and WebSocket.

## Features

- 🌐 Real-time messaging over WebSocket
- 📱 Simple HTML/JavaScript frontend
- 💾 SQLite database for message persistence
- 🔍 IP-based client identification
- 📊 Message history
- 🔄 Automatic reconnection
- 📡 Works on local WiFi networks

## Prerequisites

- Go 1.21 or higher
- SQLite3

## Installation

1. Clone the repository:
```bash
cd thamsapp
```

2. Install dependencies:
```bash
go mod download
```

## Running the Application

1. Start the server:
```bash
go run .
```

2. The server will start on port 8080. You'll see output like:
```
Server starting on http://0.0.0.0:8080
WebSocket endpoint: ws://0.0.0.0:8080/ws
```

3. Open your browser and navigate to:
```
http://<your-local-ip>:8080
```

To find your local IP:
- **Linux/Mac**: `ip addr` or `ifconfig`
- **Windows**: `ipconfig`

4. Share this URL with other devices on the same WiFi network!

## Architecture

### Backend (Go)
- **main.go**: WebSocket server and HTTP handlers
- **database.go**: SQLite database operations
- **chat.db**: SQLite database file (created automatically)

### Frontend
- **index.html**: Single-page application with inline JavaScript and CSS
- WebSocket client for real-time communication
- Responsive design

### Database Schema

**Clients Table:**
- ip (TEXT, PRIMARY KEY)
- first_seen (DATETIME)
- last_seen (DATETIME)

**Messages Table:**
- id (INTEGER, PRIMARY KEY)
- client_ip (TEXT)
- content (TEXT)
- timestamp (DATETIME)

## API Endpoints

- `GET /` - Serves the HTML frontend
- `GET /ws` - WebSocket endpoint for real-time chat
- `GET /api/clients` - Returns all connected clients
- `GET /api/messages` - Returns recent messages

## Usage

1. Each client is identified by their IP address
2. Messages are broadcasted to all connected clients in real-time
3. Message history is automatically loaded when connecting
4. All messages and client data are stored in SQLite database

## Configuration

To change the server port, modify the `port` variable in `main.go`:
```go
port := ":8080"  // Change to your desired port
```

## Security Note

This application is designed for trusted local networks only. It does not include authentication or encryption beyond what your local network provides.

## License

MIT

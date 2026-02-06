package model

import "time"

// MessageType constants
const (
	MessageTypeText  = "TEXT"
	MessageTypeJoin  = "JOIN"
	MessageTypeLeave = "LEAVE"
)

// Message represents the WebSocket message structure
type Message struct {
	UserId      string    `json:"userId"`
	Username    string    `json:"username"`
	Message     string    `json:"message"`
	Timestamp   time.Time `json:"timestamp"`
	MessageType string    `json:"messageType"`
}

// ServerResponse represents the server's response
type ServerResponse struct {
	Message
	ServerTimestamp time.Time `json:"serverTimestamp"`
	Status          string    `json:"status"` // "OK" or "ERROR"
	Error           string    `json:"error,omitempty"`
}


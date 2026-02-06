package model

import "time"

const (
	MessageTypeText  = "TEXT"
	MessageTypeJoin  = "JOIN"
	MessageTypeLeave = "LEAVE"
)

type Message struct {
	UserId      string    `json:"userId"`
	Username    string    `json:"username"`
	Message     string    `json:"message"`
	Timestamp   time.Time `json:"timestamp"`
	MessageType string    `json:"messageType"`
	RoomId      string    `json:"-"` // Not sent in JSON payload, but used for connection routing
}

type ServerResponse struct {
	Message
	ServerTimestamp time.Time `json:"serverTimestamp"`
	Status          string    `json:"status"`
	Error           string    `json:"error,omitempty"`
}


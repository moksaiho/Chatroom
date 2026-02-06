package room

import (
	"sync"

	"github.com/gorilla/websocket"
)

// Room represents a chat room with connected clients
type Room struct {
	ID          string
	Clients     map[*websocket.Conn]bool
	Register    chan *websocket.Conn
	Unregister  chan *websocket.Conn
	Broadcast   chan []byte
	mu          sync.RWMutex
}

func NewRoom(id string) *Room {
	return &Room{
		ID:         id,
		Clients:    make(map[*websocket.Conn]bool),
		Register:   make(chan *websocket.Conn),
		Unregister: make(chan *websocket.Conn),
		Broadcast:  make(chan []byte),
	}
}

func (r *Room) Run() {
	for {
		select {
		case conn := <-r.Register:
			r.mu.Lock()
			r.Clients[conn] = true
			r.mu.Unlock()
		case conn := <-r.Unregister:
			r.mu.Lock()
			if _, ok := r.Clients[conn]; ok {
				delete(r.Clients, conn)
				conn.Close()
			}
			r.mu.Unlock()
		case <-r.Broadcast:
			r.mu.RLock()
			for conn := range r.Clients {
				_ = conn // suppress unused error
				select {
				case <-make(chan bool): // Placeholder for write handling if needed
				default:
					// For assignment 1, we mostly echo back in the handler loop.
				}
			}
			r.mu.RUnlock()
		}
	}
}

// Manager manages multiple rooms
type Manager struct {
	Rooms map[string]*Room
	mu    sync.RWMutex
}

func NewManager() *Manager {
	return &Manager{
		Rooms: make(map[string]*Room),
	}
}

func (m *Manager) GetRoom(roomId string) *Room {
	m.mu.Lock()
	defer m.mu.Unlock()

	if room, ok := m.Rooms[roomId]; ok {
		return room
	}

	room := NewRoom(roomId)
	m.Rooms[roomId] = room
	go room.Run()
	return room
}

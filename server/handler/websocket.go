package handler

import (
	"chatroom/server/model"
	"chatroom/server/room"
	"encoding/json"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9]{3,20}$`)

func validateMessage(msg *model.Message) string {
	// userId validation
	uid, err := strconv.Atoi(msg.UserId)
	if err != nil || uid < 1 || uid > 100000 {
		return "userId must be between 1 and 100000"
	}

	// username validation
	if !usernameRegex.MatchString(msg.Username) {
		return "username must be 3-20 alphanumeric characters"
	}

	// message validation
	if len(msg.Message) < 1 || len(msg.Message) > 500 {
		return "message must be 1-500 characters"
	}

	// timestamp validation (checked implicitly by unmarshal, but ensure it's not zero)
	if msg.Timestamp.IsZero() {
		return "timestamp is invalid"
	}

	// messageType validation
	switch msg.MessageType {
	case model.MessageTypeText, model.MessageTypeJoin, model.MessageTypeLeave:
		// valid
	default:
		return "invalid messageType"
	}

	return ""
}

func HandleWebSocket(manager *room.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		roomId := vars["roomId"]

		if roomId == "" {
			http.Error(w, "Room ID is required", http.StatusBadRequest)
			return
		}

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println("Upgrade error:", err)
			return
		}

		room := manager.GetRoom(roomId)
		room.Register <- conn

		defer func() {
			room.Unregister <- conn
		}()

		for {
			_, p, err := conn.ReadMessage()
			if err != nil {
				log.Println("Read error:", err)
				break
			}

			var msg model.Message
			if err := json.Unmarshal(p, &msg); err != nil {
				// Invalid JSON
				response := model.ServerResponse{
					Status:          "ERROR",
					Error:           "Invalid JSON format",
					ServerTimestamp: time.Now(),
				}
				conn.WriteJSON(response)
				continue
			}

			if errStr := validateMessage(&msg); errStr != "" {
				response := model.ServerResponse{
					Message:         msg,
					Status:          "ERROR",
					Error:           errStr,
					ServerTimestamp: time.Now(),
				}
				conn.WriteJSON(response)
				continue
			}

			// Valid message - Echo back
			response := model.ServerResponse{
				Message:         msg,
				Status:          "OK",
				ServerTimestamp: time.Now(),
			}
			
			if err := conn.WriteJSON(response); err != nil {
				log.Println("Write error:", err)
				break
			}
		}
	}
}


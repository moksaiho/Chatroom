package generator

import (
	"chatroom/client-part1/model"
	"fmt"
	"math/rand"
	"time"
)

type UserState int

const (
	StateIdle UserState = iota
	StateJoined
)

var predefinedMessages = []string{
	"Hello world!", "How are you?", "WebSocket is cool", "Distributed systems are hard",
	"Java vs Go", "Chat application", "Testing high load", "Another message",
	"Design patterns", "Microservices architecture", "Latency check", "Throughput test",
	"Keep alive", "Ping", "Pong", "Good morning", "Good night", "See you later",
	"I will be back", "Connection pool", "Thread safety", "Concurrency", "Parallelism",
	"Scalability", "Reliability", "Availability", "Consistency", "Partition tolerance",
	"CAP theorem", "Little's Law", "Queuing theory", "Load balancing", "Failover",
	"Redundancy", "Replication", "Sharding", "Caching", "Database", "Network",
	"Protocol", "TCP/IP", "HTTP", "REST", "RPC", "gRPC", "JSON", "XML", "YAML",
	"Protobuf", "Message queue",
}

type Generator struct {
	TotalMessages int
	Output        chan model.Message
	userStates    map[string]UserState
	rnd           *rand.Rand
}

func NewGenerator(totalMessages int, bufferSize int) *Generator {
	return &Generator{
		TotalMessages: totalMessages,
		Output:        make(chan model.Message, bufferSize),
		userStates:    make(map[string]UserState),
		rnd:           rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (g *Generator) Run() {
	defer close(g.Output)

	for i := 0; i < g.TotalMessages; i++ {
		userIdInt := g.rnd.Intn(100000) + 1
		userId := fmt.Sprintf("%d", userIdInt)
		username := fmt.Sprintf("user%s", userId)
		roomId := fmt.Sprintf("%d", g.rnd.Intn(20)+1)

		state := g.userStates[userId]
		var msgType string

		if state == StateIdle {
			msgType = model.MessageTypeJoin
			g.userStates[userId] = StateJoined
		} else {
			// StateJoined
			r := g.rnd.Float64()
			if r < 0.05 { // 5% chance to Leave
				msgType = model.MessageTypeLeave
				g.userStates[userId] = StateIdle
			} else {
				msgType = model.MessageTypeText
			}
		}

		msgContent := predefinedMessages[g.rnd.Intn(len(predefinedMessages))]

		msg := model.Message{
			UserId:      userId,
			Username:    username,
			Message:     msgContent,
			Timestamp:   time.Now(),
			MessageType: msgType,
			RoomId:      roomId,
		}

		g.Output <- msg
	}
}


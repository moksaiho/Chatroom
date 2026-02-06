package pool

import (
	"chatroom/client-part2/metrics"
	"chatroom/client-part2/model"
	"fmt"
	"log"
	"math"
	"net/url"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type Worker struct {
	ID        int
	Input     <-chan model.Message
	Collector *metrics.Collector
	Host      string
	Conns     map[string]*websocket.Conn
	mu        sync.Mutex
}

func NewWorker(id int, input <-chan model.Message, collector *metrics.Collector, host string) *Worker {
	return &Worker{
		ID:        id,
		Input:     input,
		Collector: collector,
		Host:      host,
		Conns:     make(map[string]*websocket.Conn),
	}
}

func (w *Worker) Run(wg *sync.WaitGroup) {
	defer wg.Done()

	for msg := range w.Input {
		w.processMessageWithRetry(msg)
	}
	// Cleanup connections
	for _, conn := range w.Conns {
		conn.Close()
	}
}

func (w *Worker) getConnection(roomId string) (*websocket.Conn, error) {
	if conn, ok := w.Conns[roomId]; ok {
		return conn, nil
	}

	u := url.URL{Scheme: "ws", Host: w.Host, Path: fmt.Sprintf("/chat/%s", roomId)}
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return nil, err
	}

	w.Conns[roomId] = conn
	return conn, nil
}

func (w *Worker) processMessageWithRetry(msg model.Message) {
	maxRetries := 5
	baseDelay := 100 * time.Millisecond

	for i := 0; i <= maxRetries; i++ {
		start := time.Now()
		err := w.sendMessage(msg)
		if err == nil {
			// Success
			latency := time.Since(start).Milliseconds()
			w.Collector.Record(metrics.Record{
				Timestamp:   start,
				MessageType: msg.MessageType,
				Latency:     latency,
				StatusCode:  "OK",
				RoomId:      msg.RoomId,
			})
			return
		}

		// Failure
		log.Printf("Worker %d: Failed to send (attempt %d/%d): %v", w.ID, i+1, maxRetries+1, err)
		
		// If connection failed, close and remove it so we reconnect next time
		if conn, ok := w.Conns[msg.RoomId]; ok {
			conn.Close()
			delete(w.Conns, msg.RoomId)
		}

		if i == maxRetries {
			w.Collector.Record(metrics.Record{
				Timestamp:   start,
				MessageType: msg.MessageType,
				Latency:     0,
				StatusCode:  "ERROR",
				RoomId:      msg.RoomId,
			})
		} else {
			// Exponential backoff
			delay := baseDelay * time.Duration(math.Pow(2, float64(i)))
			time.Sleep(delay)
		}
	}
}

func (w *Worker) sendMessage(msg model.Message) error {
	conn, err := w.getConnection(msg.RoomId)
	if err != nil {
		return err
	}

	// Set deadline for write
	conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
	if err := conn.WriteJSON(msg); err != nil {
		return err
	}

	// Set deadline for read
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	
	// Read response (Echo)
	// We use RawMessage to avoid full parsing if we don't need it, but we need to check ID/Status ideally.
	// For performance test, just successfully reading a JSON is good enough proof of roundtrip.
	_, _, err = conn.ReadMessage()
	if err != nil {
		return err
	}

	return nil
}

type Pool struct {
	NumWorkers int
	GeneratorInput <-chan model.Message
	Collector  *metrics.Collector
	Host       string
}

func NewPool(numWorkers int, input <-chan model.Message, collector *metrics.Collector, host string) *Pool {
	return &Pool{
		NumWorkers: numWorkers,
		GeneratorInput: input,
		Collector:  collector,
		Host:       host,
	}
}

func (p *Pool) Run() {
	var wg sync.WaitGroup
	for i := 0; i < p.NumWorkers; i++ {
		wg.Add(1)
		worker := NewWorker(i, p.GeneratorInput, p.Collector, p.Host)
		go worker.Run(&wg)
	}
	wg.Wait()
}


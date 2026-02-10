package main

import (
	"chatroom/client-part1/generator"
	"chatroom/client-part1/metrics"
	"chatroom/client-part1/model"
	"chatroom/client-part1/pool"
	"flag"
	"fmt"
	"log"
	"net/url"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

func main() {
	host := flag.String("host", "localhost:8080", "Server host:port")
	workers := flag.Int("workers", 900, "Number of worker threads")
	totalMessages := flag.Int("messages", 500000, "Total number of messages to send")
	flag.Parse()

	fmt.Printf("Starting Client Part 1 with host=%s, workers=%d, messages=%d\n", *host, *workers, *totalMessages)

	// Warmup Phase
	fmt.Println("\n--- Starting Warmup Phase ---")
	warmupDuration := runWarmup(*host, *workers, 1000)
	fmt.Println("--- Warmup Complete ---")

	// Little's Law Analysis
	estimatedRTT := warmupDuration.Seconds() / 1000.0
	predictedThroughput := float64(*workers) / estimatedRTT

	fmt.Println("\n--- Little's Law Prediction ---")
	fmt.Printf("Workers (L): %d\n", *workers)
	fmt.Printf("Estimated RTT (W): %.5f seconds (based on warmup)\n", estimatedRTT)
	fmt.Printf("Predicted Throughput (lambda = L / W): %.2f msg/sec\n", predictedThroughput)
	fmt.Println("-------------------------------")

	// Main Phase
	fmt.Println("\n--- Starting Main Phase ---")
	
	// metrics (Simplified for Part 1)
	collector := metrics.NewCollector()
	collector.Start()

	// generator
	// Buffer size 10000 to avoid blocking generator
	gen := generator.NewGenerator(*totalMessages, 10000)
	go gen.Run()

	// pool
	p := pool.NewPool(*workers, gen.Output, collector, *host)
	
	start := time.Now()
	p.Run()
	duration := time.Since(start)

	// Shutdown collector
	collector.Close()
	<-collector.Done

	fmt.Println("--- Main Phase Complete ---")
	collector.PrintSummary()
	fmt.Printf("Wall Time: %.2f seconds\n", duration.Seconds())
}

func runWarmup(host string, numWorkers int, msgsPerWorker int) time.Duration {
	var wg sync.WaitGroup
	start := time.Now()

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			
			// Simple dial and send loop
			u := url.URL{Scheme: "ws", Host: host, Path: "/chat/1"}
			conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
			if err != nil {
				log.Printf("Warmup worker %d failed to connect: %v", id, err)
				return
			}
			defer conn.Close()

			for j := 0; j < msgsPerWorker; j++ {
				msg := model.Message{
					UserId:      "warmup",
					Username:    "warmup",
					Message:     "warmup message",
					Timestamp:   time.Now(),
					MessageType: "TEXT",
				}
				
				conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
				if err := conn.WriteJSON(msg); err != nil {
					log.Printf("Warmup worker %d failed write: %v", id, err)
					return // Stop this worker on error
				}
				
				conn.SetReadDeadline(time.Now().Add(5 * time.Second))
				_, _, err := conn.ReadMessage()
				if err != nil {
					log.Printf("Warmup worker %d failed read: %v", id, err)
					return
				}
			}
		}(i)
	}
	wg.Wait()
	duration := time.Since(start)
	fmt.Printf("Warmup finished in %.2f seconds\n", duration.Seconds())
	return duration
}

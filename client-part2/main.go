package main

import (
	"chatroom/client-part2/generator"
	"chatroom/client-part2/metrics"
	"chatroom/client-part2/model"
	"chatroom/client-part2/pool"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

func main() {
	host := flag.String("host", "localhost:8080", "Server host:port")
	workers := flag.Int("workers", 32, "Number of worker threads")
	totalMessages := flag.Int("messages", 500000, "Total number of messages to send")
	flag.Parse()

	fmt.Printf("Starting Client with host=%s, workers=%d, messages=%d\n", *host, *workers, *totalMessages)

	// Warmup Phase
	fmt.Println("\n--- Starting Warmup Phase ---")
	warmupDuration := runWarmup(*host, *workers, 1000)
	fmt.Println("--- Warmup Complete ---")

	// Little's Law Analysis
	// Estimated RTT = Total Duration / (Workers * MessagesPerWorker) roughly,
	// or better: RTT approx 0.5ms on local loopback.
	// We calculate avg RTT from warmup:
	estimatedRTT := warmupDuration.Seconds() / 1000.0
	predictedThroughput := float64(*workers) / estimatedRTT

	fmt.Println("\n--- Little's Law Prediction ---")
	fmt.Printf("Workers (L): %d\n", *workers)
	fmt.Printf("Estimated RTT (W): %.5f seconds (based on warmup)\n", estimatedRTT)
	fmt.Printf("Predicted Throughput (lambda = L / W): %.2f msg/sec\n", predictedThroughput)
	fmt.Println("-------------------------------")

	// Main Phase
	fmt.Println("\n--- Starting Main Phase ---")
	
	// Create results directory
	if err := os.MkdirAll("results", 0755); err != nil {
		log.Fatalf("Failed to create results directory: %v", err)
	}

	// metrics
	collector, err := metrics.NewCollector("results/results.csv")
	if err != nil {
		log.Fatalf("Failed to create collector: %v", err)
	}
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

	// Generate Chart
	if err := collector.GenerateChart("results/throughput_chart.html"); err != nil {
		log.Printf("Failed to generate chart: %v", err)
	} else {
		fmt.Println("Chart generated: results/throughput_chart.html")
	}
}

func runWarmup(host string, numWorkers int, msgsPerWorker int) time.Duration {
	var wg sync.WaitGroup
	start := time.Now()

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			
			// Simple dial and send loop
			// We just pick a random room (e.g., "1") for warmup
			u := url.URL{Scheme: "ws", Host: host, Path: "/chat/1"}
			conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
			if err != nil {
				log.Printf("Warmup worker %d failed to connect: %v", id, err)
				return
			}
			defer conn.Close()

			for j := 0; j < msgsPerWorker; j++ {
				msg := model.Message{
					UserId:      "123",
					Username:    "testuser",
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

package main

import (
	"flag"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var (
	url         = "ws://35.166.121.239:8080/chat/room123" // 替换你的地址
	concurrency = flag.Int("c", 1, "concurrent connections")
	totalMsgs   = flag.Int("n", 1, "messages per connection")
)

func main() {
	flag.Parse()
	fmt.Printf("Testing with %d concurrent connections...\n", *concurrency)

	var wg sync.WaitGroup
	start := time.Now()

	latencies := make(chan time.Duration, *concurrency**totalMsgs)

	for i := 0; i < *concurrency; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
            
			// 1. 建立连接 (不计入消息延迟，但计入连接开销)
			c, _, err := websocket.DefaultDialer.Dial(url, nil)
			if err != nil {
				log.Printf("Connection error: %v", err)
				return
			}
			defer c.Close()
			// use this message{"userId":"","username":"","message":"","timestamp":"0001-01-01T00:00:00Z","messageType":"","serverTimestamp":"2026-02-06T04:51:58.500114395Z","status":"ERROR","error":"Invalid JSON format"}

			msg := map[string]interface{}{
				"userId":      "123",
				"username":    "testuser",
				"message":     "Hello!",
				"messageType": "TEXT",
				"timestamp":   time.Now().Format(time.RFC3339), // 动态生成时间戳
			}

			for j := 0; j < *totalMsgs; j++ {
				// 2. 测量消息 RTT
				
				msgStart := time.Now()
				if err := c.WriteJSON(msg); err != nil {
					log.Println("Write error:", err)
					return
				}

				_, _, err := c.ReadMessage()
				if err != nil {
					log.Println("Read error:", err)
					return
				}
				
				// 记录单次延迟
				latencies <- time.Since(msgStart)
				// // print the message only once
				// if j == 0 {
				// 	fmt.Println(string(msg))
				// }
			}
		}(i)
	}

	wg.Wait()
	close(latencies)
	totalTime := time.Since(start)

	// 计算统计数据
	var totalLatency time.Duration
	var count int64
	for l := range latencies {
		totalLatency += l
		count++
	}

	avgLatency := totalLatency.Milliseconds() / count
	throughput := float64(count) / totalTime.Seconds()

	fmt.Printf("\n=== Results ===\n")
	fmt.Printf("Total Requests: %d\n", count)
	fmt.Printf("Total Time:     %v\n", totalTime)
	fmt.Printf("Avg Latency (W): %d ms\n", avgLatency)
	fmt.Printf("Throughput (λ):  %.2f req/s\n", throughput)
	
	// Little's Law 验证
	// L = λ * W
	expectedL := throughput * (float64(avgLatency) / 1000.0)
	fmt.Printf("\n=== Little's Law Check ===\n")
	fmt.Printf("Actual Concurrency (L): %d\n", *concurrency)
	fmt.Printf("Calculated L (λ * W):   %.2f\n", expectedL)
}
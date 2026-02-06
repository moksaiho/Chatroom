package metrics

import (
	"encoding/csv"
	"fmt"
	"os"
	"sync"
	"time"
)

type Record struct {
	Timestamp   time.Time
	MessageType string
	Latency     int64 // milliseconds
	StatusCode  string
	RoomId      string
}

type Collector struct {
	records    chan Record
	wg         sync.WaitGroup
	Done       chan struct{}
	csvFile    *os.File
	csvWriter  *csv.Writer
	Stats      Statistics
}

type Statistics struct {
	TotalMessages int
	SuccessCount  int
	FailCount     int
	TotalLatency  int64
	MinLatency    int64
	MaxLatency    int64
	StartTime     time.Time
	EndTime       time.Time
}

func NewCollector(filePath string) (*Collector, error) {
	file, err := os.Create(filePath)
	if err != nil {
		return nil, err
	}

	writer := csv.NewWriter(file)
	// Write header
	writer.Write([]string{"timestamp", "messageType", "latency_ms", "statusCode", "roomId"})
	writer.Flush()

	return &Collector{
		records:   make(chan Record, 10000),
		Done:      make(chan struct{}),
		csvFile:   file,
		csvWriter: writer,
		Stats: Statistics{
			MinLatency: 1<<63 - 1, // Max int64
		},
	}, nil
}

func (c *Collector) Record(r Record) {
	c.records <- r
}

func (c *Collector) Start() {
	c.Stats.StartTime = time.Now()
	go func() {
		for r := range c.records {
			c.Stats.TotalMessages++
			if r.StatusCode == "OK" {
				c.Stats.SuccessCount++
				c.Stats.TotalLatency += r.Latency
				if r.Latency < c.Stats.MinLatency {
					c.Stats.MinLatency = r.Latency
				}
				if r.Latency > c.Stats.MaxLatency {
					c.Stats.MaxLatency = r.Latency
				}
			} else {
				c.Stats.FailCount++
			}

			c.csvWriter.Write([]string{
				r.Timestamp.Format(time.RFC3339),
				r.MessageType,
				fmt.Sprintf("%d", r.Latency),
				r.StatusCode,
				r.RoomId,
			})
		}
		c.csvWriter.Flush()
		c.csvFile.Close()
		c.Stats.EndTime = time.Now()
		close(c.Done)
	}()
}

func (c *Collector) Close() {
	close(c.records)
}

func (c *Collector) PrintSummary() {
	duration := c.Stats.EndTime.Sub(c.Stats.StartTime).Seconds()
	throughput := float64(c.Stats.SuccessCount) / duration
	avgLatency := float64(c.Stats.TotalLatency) / float64(c.Stats.SuccessCount)

	fmt.Println("========= Test Results =========")
	fmt.Printf("Total Duration: %.2f seconds\n", duration)
	fmt.Printf("Total Messages: %d\n", c.Stats.TotalMessages)
	fmt.Printf("Successful: %d\n", c.Stats.SuccessCount)
	fmt.Printf("Failed: %d\n", c.Stats.FailCount)
	fmt.Printf("Throughput: %.2f msg/sec\n", throughput)
	fmt.Printf("Avg Latency: %.2f ms\n", avgLatency)
	fmt.Printf("Min Latency: %d ms\n", c.Stats.MinLatency)
	fmt.Printf("Max Latency: %d ms\n", c.Stats.MaxLatency)
	fmt.Println("================================")
}


package metrics

import (
	"fmt"
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
	Done       chan struct{}
	Stats      Statistics
}

type Statistics struct {
	TotalMessages    int
	SuccessCount     int
	FailCount        int
	TotalConnections int
	RetryCount       int
	StartTime        time.Time
	EndTime          time.Time
}

func NewCollector() *Collector {
	return &Collector{
		records:   make(chan Record, 10000),
		Done:      make(chan struct{}),
	}
}

func (c *Collector) Record(r Record) {
	c.records <- r
}

func (c *Collector) Start() {
	c.Stats.StartTime = time.Now()
	go func() {
		for r := range c.records {
			// Handle special status codes for connection/retry stats
			if r.StatusCode == "CONN_NEW" {
				c.Stats.TotalConnections++
				continue
			}
			if r.StatusCode == "RETRY" {
				c.Stats.RetryCount++
				continue
			}

			c.Stats.TotalMessages++
			if r.StatusCode == "OK" {
				c.Stats.SuccessCount++
			} else {
				c.Stats.FailCount++
			}
		}
		c.Stats.EndTime = time.Now()
		close(c.Done)
	}()
}

func (c *Collector) RecordConnection() {
	c.records <- Record{StatusCode: "CONN_NEW"}
}

func (c *Collector) RecordRetry() {
	c.records <- Record{StatusCode: "RETRY"}
}

func (c *Collector) Close() {
	close(c.records)
}

func (c *Collector) PrintSummary() {
	duration := c.Stats.EndTime.Sub(c.Stats.StartTime).Seconds()
	throughput := float64(c.Stats.SuccessCount) / duration

	fmt.Println("========= Test Results (Part 1) =========")
	fmt.Printf("Total Duration: %.2f seconds\n", duration)
	fmt.Printf("Total Messages: %d\n", c.Stats.TotalMessages)
	fmt.Printf("Successful: %d\n", c.Stats.SuccessCount)
	fmt.Printf("Failed: %d\n", c.Stats.FailCount)
	fmt.Printf("Throughput: %.2f msg/sec\n", throughput)
	fmt.Printf("Total Connections: %d\n", c.Stats.TotalConnections)
	fmt.Printf("Total Retries: %d\n", c.Stats.RetryCount)
	fmt.Println("=========================================")
}

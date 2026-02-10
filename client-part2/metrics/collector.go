package metrics

import (
	"encoding/csv"
	"fmt"
	"html/template"
	"os"
	"sort"
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
	TotalMessages    int
	SuccessCount     int
	FailCount        int
	TotalConnections int
	RetryCount       int
	TotalLatency     int64
	MinLatency       int64
	MaxLatency       int64
	StartTime        time.Time
	EndTime          time.Time

	// New fields for detailed analysis
	Latencies         []int64
	RoomCounts        map[string]int
	TypeCounts        map[string]int
	ThroughputBuckets map[int64]int // Unix timestamp (10s bucket) -> count
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
			MinLatency:        1<<63 - 1, // Max int64
			Latencies:         make([]int64, 0),
			RoomCounts:        make(map[string]int),
			TypeCounts:        make(map[string]int),
			ThroughputBuckets: make(map[int64]int),
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
				c.Stats.TotalLatency += r.Latency
				if r.Latency < c.Stats.MinLatency {
					c.Stats.MinLatency = r.Latency
				}
				if r.Latency > c.Stats.MaxLatency {
					c.Stats.MaxLatency = r.Latency
				}

				// Collect detailed stats
				c.Stats.Latencies = append(c.Stats.Latencies, r.Latency)
				c.Stats.RoomCounts[r.RoomId]++
				c.Stats.TypeCounts[r.MessageType]++

				// 10-second buckets
				bucket := r.Timestamp.Unix() / 10 * 10
				c.Stats.ThroughputBuckets[bucket]++

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

func (c *Collector) RecordConnection() {
	c.records <- Record{StatusCode: "CONN_NEW"}
}

func (c *Collector) RecordRetry() {
	c.records <- Record{StatusCode: "RETRY"}
}

func (c *Collector) Close() {
	close(c.records)
}

func (c *Collector) CalculatePercentiles() (median, p95, p99 int64) {
	if len(c.Stats.Latencies) == 0 {
		return 0, 0, 0
	}
	sort.Slice(c.Stats.Latencies, func(i, j int) bool {
		return c.Stats.Latencies[i] < c.Stats.Latencies[j]
	})

	n := len(c.Stats.Latencies)
	median = c.Stats.Latencies[n/2]
	p95 = c.Stats.Latencies[int(float64(n)*0.95)]
	p99 = c.Stats.Latencies[int(float64(n)*0.99)]
	return
}

func (c *Collector) PrintSummary() {
	duration := c.Stats.EndTime.Sub(c.Stats.StartTime).Seconds()
	throughput := float64(c.Stats.SuccessCount) / duration
	avgLatency := float64(c.Stats.TotalLatency) / float64(c.Stats.SuccessCount)
	median, p95, p99 := c.CalculatePercentiles()

	fmt.Println("========= Test Results =========")
	fmt.Printf("Total Duration: %.2f seconds\n", duration)
	fmt.Printf("Total Messages: %d\n", c.Stats.TotalMessages)
	fmt.Printf("Successful: %d\n", c.Stats.SuccessCount)
	fmt.Printf("Failed: %d\n", c.Stats.FailCount)
	fmt.Printf("Throughput: %.2f msg/sec\n", throughput)
	fmt.Printf("Total Connections: %d\n", c.Stats.TotalConnections)
	fmt.Printf("Total Retries: %d\n", c.Stats.RetryCount)
	fmt.Printf("Avg Latency: %.2f ms\n", avgLatency)
	fmt.Printf("Min Latency: %d ms\n", c.Stats.MinLatency)
	fmt.Printf("Max Latency: %d ms\n", c.Stats.MaxLatency)
	fmt.Printf("Median Latency: %d ms\n", median)
	fmt.Printf("P95 Latency: %d ms\n", p95)
	fmt.Printf("P99 Latency: %d ms\n", p99)
	
	fmt.Println("\n--- Message Type Distribution ---")
	for k, v := range c.Stats.TypeCounts {
		fmt.Printf("%s: %d\n", k, v)
	}
	
	fmt.Println("================================")
}

func (c *Collector) GenerateChart(filename string) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	// Prepare data for chart
	var labels []string
	var data []int

	// Sort buckets by time
	var buckets []int64
	for k := range c.Stats.ThroughputBuckets {
		buckets = append(buckets, k)
	}
	sort.Slice(buckets, func(i, j int) bool { return buckets[i] < buckets[j] })

	for _, b := range buckets {
		labels = append(labels, time.Unix(b, 0).Format("15:04:05"))
		// Throughput = count / 10 seconds
		data = append(data, c.Stats.ThroughputBuckets[b]/10)
	}

	const tpl = `
<!DOCTYPE html>
<html>
<head>
    <title>Throughput Chart</title>
    <script src="https://cdn.jsdelivr.net/npm/chart.js"></script>
</head>
<body>
    <div style="width: 80%; margin: auto;">
        <canvas id="myChart"></canvas>
    </div>
    <script>
        const ctx = document.getElementById('myChart').getContext('2d');
        const myChart = new Chart(ctx, {
            type: 'line',
            data: {
                labels: {{.Labels}},
                datasets: [{
                    label: 'Throughput (msg/sec)',
                    data: {{.Data}},
                    borderColor: 'rgb(75, 192, 192)',
                    tension: 0.1
                }]
            },
            options: {
                scales: {
                    y: {
                        beginAtZero: true
                    }
                }
            }
        });
    </script>
</body>
</html>`

	t, err := template.New("chart").Parse(tpl)
	if err != nil {
		return err
	}

	return t.Execute(f, struct {
		Labels []string
		Data   []int
	}{
		Labels: labels,
		Data:   data,
	})
}


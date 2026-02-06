# WebSocket Client (Part 2 - Performance Analysis)

This is the advanced version of the load testing client, corresponding to **Part 3: Performance Analysis** of the assignment.

## Features

- **Full Multithreading**: Connection pooling with configurable worker threads.
- **Message Generation**: State-machine based message generation (JOIN -> TEXT -> LEAVE).
- **Advanced Metrics**:
  - Detailed latency statistics (Mean, Median, P95, P99).
  - Throughput calculation.
- **CSV Export**: Writes per-message metrics to `results.csv`.

## Usage

```bash
# Build
go build -o client_part2 main.go

# Run
./client_part2 -host localhost:8080 -workers 32 -messages 500000
```

## Output

The program will output a summary to the console and generate `results.csv` with the following columns:
`timestamp, messageType, latency_ms, statusCode, roomId`


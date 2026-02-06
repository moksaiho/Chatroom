# WebSocket Client (Part 1 - Basic Load Testing)

This is the basic version of the load testing client, corresponding to **Part 2: Build the Multithreaded WebSocket Client** of the assignment.

## Features

- **Multithreading**: Implements connection pooling and worker threads.
- **Warmup Phase**: Initial 32 threads sending 1000 messages each.
- **Main Phase**: High-volume message generation (500,000 messages).
- **Basic Metrics**: Reports Total Sent, Failed, Runtime, and Throughput.

## Usage

```bash
# Build
go build -o client_part1 main.go

# Run
./client_part1 -host localhost:8080 -workers 32 -messages 500000
```

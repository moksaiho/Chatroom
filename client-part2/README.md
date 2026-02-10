# WebSocket Client (Part 2 - Performance Analysis)


## Usage

```bash
# Build
go build -o client_part2 main.go

# Run
./client_part2 -host localhost:8080 -workers 32 -messages 500000
```

## Output

The program will output a summary to the console and generate `results.csv` a html file in results folder to show the throughput in each second


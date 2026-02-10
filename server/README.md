# WebSocket Chat Server

Implementation of the distributed chat server (Part 1).

## Features

- **WebSocket Endpoint**: `/chat/{roomId}` for real-time messaging.
- **Health Check**: `/health` endpoint.
- **Room Management**: In-memory management of chat rooms and connections.
- **Validation**: Strict validation of incoming message JSON.

## Running Locally

```bash
# Build
go build -o server main.go

# Run
./server
```

Server starts on port `8080`.

## Deployment on AWS EC2

1. Launch an EC2 instance (Amazon Linux 2) and generate the key for making connection.

chmod 400 your-key.pem
ssh -i your-key.pem ec2-user@YOUR_EC2_PUBLIC_IP

2. set up the environment in ec2
sudo yum update -y
wget https://go.dev/dl/go1.21.5.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.21.5.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc
go version
if the command gives you the version of go then the go environment is set up


3. pull the code from github and run it on the server
sudo yum install git -y
git clone https://github.com/moksaiho/Chatroom.git 
cd ~/Chatroom/server
go mod download
go build -o chatroom-server main.go
./chatroom-server



4. test the running using wscat or curl

5. other userful command
 sudo systemctl daemon-reload
 sudo systemctl start chatroom
 sudo systemctl status chatroom

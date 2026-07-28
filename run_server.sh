echo "Building and running server";
GOOS=linux GOARCH=amd64 go build -o server server.go
./server

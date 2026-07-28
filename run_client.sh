echo "Building and running client";
GOOS=linux GOARCH=amd64 go build -o client client.go
./client

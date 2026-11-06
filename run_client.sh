echo "Building and running client";
GOOS=linux GOARCH=amd64 go build -o client main_infra/client/client.go
echo "Client needs user inputs. Run it with ./client -target"
#./client

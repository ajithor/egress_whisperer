#echo 'if you dont have server.crt and server.key, run \nopenssl req -x509 -newkey rsa:4096 -keyout server.key -out server.crt -sha256 -days 365 -nodes -subj "/CN=egress_whisperer.local"'
echo "Building and running server";
GOOS=linux GOARCH=amd64 go build -o server main_infra/server/server.go
./server

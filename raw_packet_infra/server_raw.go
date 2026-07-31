package main
import (
	"fmt"
	"net"
	"log"
	"sync"
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"
	"crypto/tls"

)

// Create a global pool of byte slices to reuse memory across goroutines, in case the project goes to a 65k port option
var bufPool = sync.Pool{
	New: func() interface{} {
		return make([]byte, 1024)
	},
}

const c2_server_ip = "0.0.0.0"

var http_response_string = "HTTP/1.1 200 OK\r\n" +
	"Date: Wed, 29 Jul 2026 22:30:00 GMT\r\n" +
	"Server: Apache/2.4.41 (Ubuntu)\r\n" +
	"Content-Type: text/html; charset=UTF-8\r\n" +
	"Content-Length: 42\r\n" +
	"Connection: close\r\n" +
	"\r\n" +
	"<html><body><h1>It Works!</h1></body></html>"

func invokeTlsListener(ctx context.Context, port string){
	cert, err := tls.LoadX509KeyPair("server.crt", "server.key")
	if err != nil {
		log.Fatalf("[-] Failed to load TLS certificates: %v\n", err)
		<-ctx.Done()
	}

	config := &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12, // Ensure compliance with modern security baselines
	}

	//Open the network transport layer port, i.e initiate tsl listener
	listener, err := tls.Listen("tcp", c2_server_ip+":"+port, config)
	if err != nil {
		log.Fatalf("[-] Failed to initialize TLS listener: %v\n", err)
		<-ctx.Done()
	}
	defer listener.Close()

	for {
		// The tls.Listen handles uConn.Handshake() automatically.
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("[-] Handshake validation failed: %v\n", err)
			continue
		}
		
		go func(c net.Conn) {
			defer c.Close()
			
			buf := bufPool.Get().([]byte)
			defer bufPool.Put(buf)

			conn.SetDeadline(time.Now().Add(3 * time.Second)) //set deadlines to aoid tarpits and other io timeouts
			// decrypted HTTP payload
			n, err := c.Read(buf)
			if err!=nil{
				log.Println(err)
			}
			fmt.Printf("[*] Received inner application payload over TLS port %q\n%s\n", port, string(buf[:n]))

			c.Write([]byte(http_response_string))
		}(conn)
	}
}

func invokeTcpListener(ctx context.Context, port string){

	listener, err := net.Listen("tcp", c2_server_ip+":"+port) //bind port to listener
	if err!=nil{
		log.Println(err) //usually when IP is wrong or port is busy
		return
	}
	defer listener.Close();
	go func(){
		<-ctx.Done()
		listener.Close()
	}()

	for{
		//wait forever and whenever we get a connection, once we do, 
		conn, err := listener.Accept()
		if err!=nil{
			select{
			case <-ctx.Done():
					return
			default:
				log.Println(err)
				continue
			}
		}
		go handleConnections(conn)

	}
}

func invokeUdpListener(ctx context.Context, port string){

	conn, err := net.ListenPacket("udp", c2_server_ip+":"+port)
	if err != nil {
    	log.Println(err)
    	<-ctx.Done()
    	return
	}
	defer conn.Close()
	go func(){
		<-ctx.Done()
		conn.Close()
	}()

	// Borrow a buffer from the pool
	buf := bufPool.Get().([]byte)
	// Return to the pool when the function exits
	defer bufPool.Put(buf)
	for{
		n, clientAddr, err := conn.ReadFrom(buf)
		if err != nil {
			if ctx.Err() != nil {	//error occured because the context closed
				return
			}
        	continue
    	}
		fmt.Println("echo -", string(buf[:n]))
		conn.WriteTo([]byte(buf[:n]), clientAddr)
	} 

}

func handleConnections(conn net.Conn){
	defer conn.Close()
	// Borrow a buffer from the pool
	buf := bufPool.Get().([]byte)
	// Returned to the pool when the function exits
	defer bufPool.Put(buf)

	conn.SetDeadline(time.Now().Add(3 * time.Second)) //set deadlines to aoid tarpits and other io timeouts
	n, err := conn.Read(buf)
	if err!=nil{
		log.Println(err)
	}
	
	fmt.Println("echo - ", string(buf[:n]))
	//time.Sleep(10*time.Second) //to test the read timeout on the client
	
	fmt.Fprintf(conn, "Hello, a beacon")
}

func main() {
	fmt.Println("server started\n")
	port_list := []string{"80", "8080", "443", "53", "123", "9001", "8443"}
	//firewall port-filtering rules are set to accept the above listed ports, except 9001 - stage 1

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	
	for _, port := range port_list{
		wg.Add(1)
		go func (p string){
			defer wg.Done()
			switch p {
			 case "443", "80", "8080", "9001":
			 	go invokeTcpListener(ctx, p)
			 case "84430":
			 	go invokeTlsListener(ctx, p)
			 default:
			 	go invokeUdpListener(ctx, p)
			}
			//if we use "port here instead of p, reference gets passed, so all parallel funtionas may change it"
		}(port)
	}
	//graceful exit on ctrl+c
	// wait for Ctrl+C / SIGTERM
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	<-sigCh

	fmt.Println("shutting down...")
	cancel()   // tell listeners to stop
	wg.Wait()  // wait for them to actually close
	fmt.Println("shutdown complete")
}

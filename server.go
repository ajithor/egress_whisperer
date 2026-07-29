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

)

// Create a global pool of byte slices to reuse memory across goroutines, in case the project goes to a 65k port option
var bufPool = sync.Pool{
	New: func() interface{} {
		return make([]byte, 1024)
	},
}

func invokeTcpListener(ctx context.Context, port string){

	listener, err := net.Listen("tcp", "10.0.0.1:"+port) //bind port to listener
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

	conn, err := net.ListenPacket("udp", "10.0.0.1:"+port)
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
		fmt.Println("buffer", string(buf[:n]))
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
	
	fmt.Println("buffer", string(buf[:n]))
	//time.Sleep(10*time.Second) //to test the read timeout on the client
	fmt.Fprintf(conn, "Echo - "+ string(buf[:n]) )
}

func main() {
	fmt.Println("This is server code")
	port_list := []string{"80", "8080", "443", "53", "123", 9001}
	//firewall port-filtering rules are set to accept the above listed ports, except 9001 - stage 1

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	
	for _, port := range port_list{
		wg.Add(1)
		go func (p string){
			defer wg.Done()
			if p=="53" || p=="123"{
				go invokeUdpListener(ctx, p)
			}else{
				go invokeTcpListener(ctx, p)
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

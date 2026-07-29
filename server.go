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

)

func invokeTcpListener(ctx context.Context, port string){

	listener, err := net.Listen("tcp", "127.0.0.1:"+port) //bind port to listener
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

func handleConnections(conn net.Conn){
	defer conn.Close()
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err!=nil{
		log.Println(err)
	}
	
	fmt.Println("buffer", string(buf[:n]))
	fmt.Fprintf(conn, "Echo - "+ string(buf[:n]) )
}

func main() {
	fmt.Println("This is server code")
	port_list := []string{"80", "8080", "443", "53", "123"}
	
	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	
	for _, port := range port_list{
		wg.Add(1)
		go func (p string){
			defer wg.Done()
			if p=="53" || p=="123"{
				fmt.Println("Invoking udp listener on port "+p)
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
package main
import (
	"fmt"
	"net"
	"log"

)

func invokeParallelListeners(port string){

	fmt.Println("binding 127.0.0.1:"+port)
	listener, err := net.Listen("tcp", "127.0.0.1:"+port) //bind port to listener
	if err!=nil{
		log.Println(err) //usually when IP is wrong or port is busy
		return
	}
	defer listener.Close();
	
	fmt.Println("Listener bound to ", listener.Addr())

	for{
		//wait forever and whenever we get a connection, once we do, 
		conn, err := listener.Accept()
		if err!=nil{
			log.Println(err)
			continue
		}
		go handleConnections(conn)

	}
}

func handleConnections(conn net.Conn){
	defer conn.Close()
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err!=nil{
		fmt.Println(err)
	}
	fmt.Println("buffer", string(buf[:n]))

	fmt.Fprintf(conn, "Echo - "+ string(buf[:n]) )
}

func main() {
	fmt.Println("This is server code")
	//port_list := []string{"80", "8080", "443"} 
	//for _, port := range port_list{
		port :="80"
		fmt.Println("Invoking listener for ", port)
		go invokeParallelListeners(port)
	//}
	select{}
}
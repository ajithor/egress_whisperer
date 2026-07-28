package main

import (
	"fmt"
	"net"
)
func invokeParallelDialups(port string){
	
	conn, err := net.Dial("tcp", "127.0.0.1:"+port)
	if err!=nil{
		fmt.Println("bad connection ", err)
	}
	fmt.Println("Sent dialup to server at"+port+" with handle ", conn)

	fmt.Fprintf(conn, "Hey, beacon message! on port "+port)
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err!=nil{
		fmt.Println(err)
	}
	fmt.Println("Recieved foll msg from server\n", string(buf[:n]))
	defer conn.Close()
}

func main(){
	fmt.Println("Client code")

	port_list := []string{"80", "8080", "443"} 
	for _, port := range port_list{
		fmt.Println("Invoking listener for ", port)
		go invokeParallelDialups(port)
	}

	select{}
}
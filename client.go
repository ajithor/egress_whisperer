package main

import (
	"fmt"
	"net"
)

func main(){
	fmt.Println("Client code")
	conn, err := net.Dial("tcp", "127.0.0.1:80")
	if err!=nil{
		fmt.Println("bad connection ", err)
	}
	fmt.Println("Sent dialup to server at 80 with handle ", conn)
	fmt.Fprintf(conn, "Hey, beacon message!")
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err!=nil{
		fmt.Println(err)
	}
	fmt.Println("Recieved foll msg from server\n", string(buf[:n]))
	defer conn.Close()
}
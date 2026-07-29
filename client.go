package main

import (
	"fmt"
	"net"
	"time"
	"log"
)

type dialResult struct {
	port   string
	status string
}

func invokeParallelDialups(port string, resultChan chan<- dialResult){
	
	conn, err := net.DialTimeout("tcp", "127.0.0.1:"+port, 2*time.Second)
	if err!=nil{
		log.Println(err)
		resultChan <- dialResult{port: port, status: err.Error()}
		return
	}
	defer conn.Close()

	fmt.Fprintf(conn, "Hey, beacon message! on port "+port)
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err!=nil{
		log.Println(err)
		resultChan <- dialResult{port: port, status: err.Error()}
	}
	fmt.Println("Recieved foll msg from server -", string(buf[:n]))
	resultChan <- dialResult{port: port, status: "success"}
}

func main(){
	fmt.Println("Client code")
	port_list := []string{"80", "8080", "443", "9001"} // ,"53", "123"}
	report := make(map[string]string)

	resultChan := make(chan dialResult, len(port_list))

	for _, port := range port_list{
		if port=="53"||port=="123"{
			fmt.Println("Invoking udp dialup for port "+port)
		}else{
			go invokeParallelDialups(port, resultChan)
		}
	}

	for i := 0; i < len(port_list); i++ {
		res := <-resultChan
		report[res.port] = res.status // Only main routine writes to map (Safe!)
	}

	fmt.Printf("---Final Report---\n")
	for key, val := range report{
		fmt.Println(key,"\t:\t", val)
	}
}
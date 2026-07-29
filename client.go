package main

import (
	"fmt"
	"net"
	"time"
	"log"
	"sync"
)

// Create a global pool of byte slices to reuse memory across goroutines, in case the project goes to a 65k port option
var bufPool = sync.Pool{
	New: func() interface{} {
		return make([]byte, 1024)
	},
}

type dialResult struct {
	port   string
	status string
}

func invokeParallelDialups(port string, wg *sync.WaitGroup, resultChan chan<- dialResult){
	
	defer wg.Done()

	var con_type string
	if port=="53"||port=="123"{ con_type="udp" } else {con_type ="tcp"}
	conn, err := net.DialTimeout(con_type, "10.0.0.1:"+port, 2*time.Second)
	if err!=nil{
		log.Println(err)
		//fmt.Println("port "+port+" coming mark 1")
		resultChan <- dialResult{port: port, status: err.Error()}
		return
	}
	defer conn.Close()

	conn.SetDeadline(time.Now().Add(3 * time.Second))
	fmt.Fprintf(conn, "Hey, beacon message! on port "+port)
	
	// Borrow a buffer from the pool
	buf := bufPool.Get().([]byte)
	// Return to the pool when the function exits
	defer bufPool.Put(buf)

	n, err := conn.Read(buf)
	if err!=nil{
		log.Println(err)
		//fmt.Println("port "+port+" coming mark 2")
		resultChan <- dialResult{port: port, status: err.Error()}
		return
	}
	fmt.Println("Recieved foll msg from server -", string(buf[:n]))
	//fmt.Println("port "+port+" coming mark 3")
	resultChan <- dialResult{port: port, status: "success"}
}

func main(){
	fmt.Println("Client code")
	port_list := []string{"80", "8080", "443", "53", "123", "9001"}
	//firewall port-filtering rules are set to accept the above listed ports, except 9001 - stage 1
	
	report := make(map[string]string)

	resultChan := make(chan dialResult, len(port_list))
	var wg sync.WaitGroup

	for _, port := range port_list{
		wg.Add(1)
		go invokeParallelDialups(port, &wg, resultChan)
	}

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	for i := 0; i < len(port_list); i++ {
		res := <-resultChan
		report[res.port] = res.status // Only main routine writes to map (Safe!)
	}

	fmt.Printf("---Final Report---\n")
	for key, val := range report{
		fmt.Println(key,"\t:\t", val)
	}
}

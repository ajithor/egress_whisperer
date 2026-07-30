package main

import (
	"fmt"
	"net"
	"time"
	"log"
	"sync"
	"strings"
	"flag"
	utls "github.com/refraction-networking/utls"
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

var c2_server_ip string
var timeout time.Duration

var http_request_string = "GET /index.html HTTP/1.1\r\n" +
           "Host: " + c2_server_ip + "\r\n" +
           "User-Agent: Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36\r\n" +
           "Accept: text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8\r\n" +
           "Accept-Language: en-US,en;q=0.5\r\n" +
           "Connection: close\r\n" + 
           "\r\n"

func invokeParallelDialups(port string, wg *sync.WaitGroup, resultChan chan<- dialResult){
	
	defer wg.Done()

	var con_type string
	if port=="53"||port=="123"{ con_type="udp" } else {con_type ="tcp"}

	conn, err := net.DialTimeout(con_type, c2_server_ip+":"+port, timeout)
	if err!=nil{
		log.Println(err)
		//fmt.Println("port "+port+" coming mark 1")
		resultChan <- dialResult{port: port, status: err.Error()}
		return
	}
	defer conn.Close()

	conn.SetDeadline(time.Now().Add(timeout))

	switch port {
	case "443", "8443":
		config := &utls.Config{InsecureSkipVerify: true}
		uConn := utls.UClient(conn, config, utls.HelloChrome_Auto)
		uConn.SetDeadline(time.Now().Add(timeout))

		err = uConn.Handshake()
		if err != nil {
			log.Println("TLS Handshake Blocked or Failed:", err)
			resultChan <- dialResult{port: port, status: err.Error()}
			return
		}
		defer uConn.Close()

		fmt.Fprintf(uConn, http_request_string)
		
		// Borrow a buffer from the pool
		buf := bufPool.Get().([]byte)
		// Return to the pool when the function exits
		defer bufPool.Put(buf)

		n, err := uConn.Read(buf)
		if err!=nil{
			log.Println(err)
			//fmt.Println("port "+port+" coming mark 6")
			resultChan <- dialResult{port: port, status: err.Error()}
			return
		}
		
		if strings.Contains(string(buf[:n]), "200 OK"){
			resultChan <- dialResult{port: port, status: "success for "+port}
		}else{
			resultChan <- dialResult{port: port, status: "https did not return 200"}
		}

	case "80", "8080", "9001":
		fmt.Fprintf(conn, http_request_string)
		
		// Borrow a buffer from the pool
		buf := bufPool.Get().([]byte)
		// Return to the pool when the function exits
		defer bufPool.Put(buf)

		n, err := conn.Read(buf)
		if err!=nil{
			log.Println(err)
			//fmt.Println("port "+port+" coming mark 5")
			resultChan <- dialResult{port: port, status: err.Error()}
			return
		}
		
		if strings.Contains(string(buf[:n]), "200 OK"){
			resultChan <- dialResult{port: port, status: "success for "+port}
		}else{
			resultChan <- dialResult{port: port, status: "http did not return 200"}
		}

	default:
		//fmt.Println("debug for port ------", port)
		fmt.Fprintf(conn, "beacon from port : "+ port)
	
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
		fmt.Println("Recieved foll msg from server - from port",port, "\n" ,string(buf[:n]))
		//fmt.Println("port "+port+" coming mark 3")
		resultChan <- dialResult{port: port, status: "success"}
		return
	}
	
}

func main(){
	//user input collection
	target  := flag.String("target", "", "C2 server IP address (required)")
    timeout = *flag.Duration("timeout", 3*time.Second, "probe timeout per port")
    ports   := flag.String("ports", "80,8080,443,53,123", "comma separated ports to probex")
    
    flag.Parse()

    if *target == "" {
        fmt.Println("Error: -target is required")
        flag.Usage()
        return
    }
    c2_server_ip = *target
    port_list := strings.Split(*ports, ",")
	for i, p := range port_list {
    	port_list[i] = strings.TrimSpace(p)
	}

	//port_list := []string{"80", "8080", "443", "53", "123", "9001", "8443"}
	//supported ports from the server-side
	
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

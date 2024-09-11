package main

import (
	"fmt"
	"log"
	functions "netcat/functions"
	"os"
	"strconv"
)

func main() {
	var port string
	if len(os.Args) == 1 {
		port = ":8989"
	} else if len(os.Args) ==2 {
		port = os.Args[1]
		portNum, err := strconv.Atoi(port)
		if err != nil || portNum < 1 || portNum > 65535 {
			fmt.Println("[USAGE]: ./TCPChat $port")
			return
		}
		port = ":" + port
	}else{
		fmt.Println("[USAGE]: ./TCPChat $port")
		return
	}
	fmt.Println("Listening on port", port)

	server := functions.NewServer("0.0.0.0" + port)
	go server.HandleConnections()

	// Start the server without WaitGroup
	if err := server.Start(); err != nil {
		log.Fatal("Server Error: ", err)
	}
}

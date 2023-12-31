package main

import (
	"flag"
	"fmt"
	"net"
	"os"
)

var directoryArg string

func handleConnection(conn net.Conn) {
	defer conn.Close()

	buffer := make([]byte, 1024)

	// read from connection
	size, err := conn.Read(buffer)
	if err != nil {
		fmt.Println("error reading from connection: ", err.Error())
	}

	request := ParseRequest(buffer[:size])
	response := CreateResponse(request)

	_, err = conn.Write(response.Encode())
	if err != nil {
		fmt.Println("error writing response: ", err.Error())
	}
}

func main() {
	flag.StringVar(&directoryArg, "directory", "", "directory where file is stored")
	flag.Parse()

	fmt.Printf("server started with directory argument: %s\n", directoryArg)

	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}
	defer l.Close()

	fmt.Println("server listening...")

	for {
		// Accept incoming connections
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			continue
		}

		// Handle the connection in a new goroutine
		go handleConnection(conn)
	}

}

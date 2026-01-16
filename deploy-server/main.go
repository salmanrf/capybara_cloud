package main

import (
	"fmt"
	"log"
	"net"
)

func NewListener() {
	listener, err := net.Listen("tcp4", "127.0.0.1:2020")

	if err != nil {
		log.Fatal("Failed to initiate socket")
	}

	defer listener.Close()

	fmt.Printf("bound to %q\n", listener.Addr())

	for {
		conn, err := listener.Accept()

		if err != nil {
			log.Fatalln("Unable to accept connection", err)
		}

		
		go func(conn net.Conn) {
			defer conn.Close()

			fmt.Println("New connection has been opened")
		}(conn)
	}
}

func main() {
	NewListener()
}
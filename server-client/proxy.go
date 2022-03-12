package main

import (
	"io"
	"log"
	"net"
)

/*
 * Proxy connection from client to server
 */
func handle(src net.Conn) {
	dst, err := net.Dial("tcp", "0.0.0.0:8080")
	if err != nil {
		log.Fatalln("Unable to connect to server from proxy")
	}
	defer dst.Close()

	// Run go routine to prevent io.Copy from blocking
	go func() {
		// Copy source's output to the destination
		log.Println("Copying data from client to server")
		if _, err := io.Copy(dst, src); err != nil {
			log.Fatalln("Failed to copy information from client to server")
		}
	}()
	// Copy destination's output back to the source
	log.Println("Copying data from server to client")
	if _, err := io.Copy(src, dst); err != nil {
		log.Fatalln("Failed to copy server information to  client")
	}
}

func main() {
	// forward/copy over to server connection
	// server is listening on 8080, port forward can listen on 8443
	listener, err := net.Listen("tcp", ":8443")
	if err != nil {
		log.Fatalln("Unable to bind port for proxy")
	}
	log.Println("Listening for client relays on 0.0.0.0:8443")
	// take input from client connection
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatalln("Unable to accept proxy connection")
		}
		go handle(conn)
	}

	// take response from server

	// forward/copy over to client

}

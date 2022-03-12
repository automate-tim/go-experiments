package main

import (
	"bytes"
	"io"
	"log"
	"net"
)

// echo is a handler function that simply echoes received data.
func driver_echo(conn net.Conn, commands chan string, results chan string) {
	defer conn.Close()

	// Create a buffer to store received data.
	b := make([]byte, 512)
	welcomeString := "Welcome to the server, run commands after the >\n"
	log.Println("Writing data to driver")
	if _, err := conn.Write([]byte(welcomeString)); err != nil {
		log.Fatalln("Unable to write data for driver welcome string")
	}
	for {
		// Receive data via conn.Read into a buffer.
		//fmt.Print(">")
		log.Println("Writing char > to driver")
		if _, err := conn.Write([]byte(">")); err != nil {
			log.Fatalln("Unable to write char > for driver")
		}
		size, err := conn.Read(b[0:])
		if err == io.EOF {
			log.Println("Driver disconnected")
			break
		}
		if err != nil {
			log.Println("Unexpected error in driver echo")
			break
		}
		go func() {
			commands <- string(b[:size])
		}()
		log.Printf("Driver action received %d bytes: %s", size, string(b))

		//if _, err := io.Copy(conn, conn); err != nil {
		//	log.Fatalln("Unable to read/write driver data")
		//}

		//testcmd := "whoami"
		//testcmddata := []byte(testcmd)
		// Send data via conn.Write.
		myresults := <-results
		resultdata := []byte(myresults)
		if resultdata != nil {
			log.Println("Writing driver response data")
			if _, err := conn.Write(resultdata); err != nil {
				log.Fatalln("Unable to write data for driver data")
			}
		}

		//b = nil
		//if _, err := conn.Write(b[0:size]); err != nil {
		//	log.Fatalln("Unable to write data")
		//}
	}
}

func sendData(conn net.Conn, cmdsIn chan string, results chan string) {
	defer conn.Close()

	// Create a buffer to store received data.
	b := make([]byte, 512)
	for {
		//testcmd := "whoami"
		testcmds := <-cmdsIn
		testcmddata := []byte(testcmds)
		// Send data via conn.Write.
		if testcmddata != nil {
			log.Println("Writing data")
			if _, err := conn.Write(testcmddata); err != nil {
				log.Fatalln("Unable to write data to client: ", err)
			}
			// Receive data via conn.Read into a buffer.
			size, err := conn.Read(b[0:])

			if err == io.EOF {
				log.Println("Client disconnected")
				break
			}
			if err != nil {
				log.Println("Unexpected error")
				break
			}
			withoutNull := bytes.Trim(b, "\x00")
			results <- string(withoutNull)
			log.Printf("sendData received %d bytes", size)
		}

	}
}

// clients will be calling back and interacting with this interface
func clientCallback(cmds chan string, results chan string) {
	// Bind to TCP port 8080 on all interfaces.
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatalln("Unable to bind to client listening port")
	}
	log.Println("Listening for clients on 0.0.0.0:8080")
	for {
		// Wait for connection. Create net.Conn on connection established.
		conn, err := listener.Accept()
		log.Println("Received client connection")
		if err != nil {
			log.Fatalln("Unable to accept client connection")
		}
		// Handle the connection. Using goroutine for concurrency.
		//go echo(conn)
		go sendData(conn, cmds, results)
	}
}

// Driver is the connection that is passing in what will be executed on any clients
func driverListen(cmds chan string, results chan string) {
	// Bind to TCP port 5050 on all interfaces.
	listener, err := net.Listen("tcp", ":5050")
	if err != nil {
		log.Fatalln("Unable to bind to driving listening port")
	}
	log.Println("Listening for drivers on 0.0.0.0:5050")
	for {
		// Wait for connection. Create net.Conn on connection established.
		conn, err := listener.Accept()
		log.Println("Received driver connection")
		if err != nil {
			log.Fatalln("Unable to accept driver connection")
		}
		// Handle the connection. Using goroutine for concurrency.
		go driver_echo(conn, cmds, results)
	}
}

func main() {
	cmdsIn := make(chan string)
	results := make(chan string)
	go clientCallback(cmdsIn, results)
	driverListen(cmdsIn, results)
}

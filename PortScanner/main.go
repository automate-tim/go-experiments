// main.go file
package main

import (
	"flag"
	"fmt"
	"net"
	"sort"
	"strconv"
	"strings"
)

const (
	porterrmsg = "Invalid port specification"
)

var verbose bool

func portScan(hostname string, ports, results chan int) {
	for port := range ports {
		hostport := fmt.Sprint(hostname, ":", port)
		conn, err := net.Dial("tcp", hostport)
		if err == nil {
			fmt.Println("Connection Successful: ", hostport)
			conn.Close()
			fmt.Println(port)
			results <- port
			continue
		}
		if verbose {
			fmt.Println("Failed: ", hostport)
		}
		results <- 0
		continue
	}
}

// Creates a slice with a range of numbers that are passed in from the minimum to the maximum
func makeRange(min, max int) []int {
	a := make([]int, max-min+1)
	for i := range a {
		a[i] = min + i
	}
	return a
}

func parsePorts(portsStr string) (ports []int) {
	portsSplit := strings.Split(portsStr, ",")
	for _, port_char := range portsSplit {
		// Checking for the simple solution first, it is a standalone int
		if nowInt, err := strconv.Atoi(port_char); err == nil {
			ports = append(ports, nowInt)
			continue
		}
		// Now Checking for range indicator -
		if strings.Contains(port_char, "-") {
			rangeSplit := strings.Split(port_char, "-")
			var rangeMinMax []int

			// Parse and make ints
			for _, i := range rangeSplit {
				j, err := strconv.Atoi(i)
				if err != nil {
					panic(err)
				}
				rangeMinMax = append(rangeMinMax, j)
			}
			// Check that range is not negative
			if rangeMinMax[1]-rangeMinMax[0] < 0 {
				println("Port range not formatted correctly, skipping")
				continue
			}
			finalRange := makeRange(rangeMinMax[0], rangeMinMax[1])
			ports = append(ports, finalRange...)
		}
	}
	return
}

/*
	Command line arguments to support for port scanning
	String hostname (scanme.nmap.org, 192.168.1.1, 127.0.0.1)
	Int port (50), port range by comma (22,80,5000), port range by hyphen (1-1000)
	Bool verbose (print fails too) (optional later)
*/
func parseCmdLineInput() (hostname string, ports []int, threads int) {
	hostnamePtr := flag.String("host", "scanme.nmap.org", "Hostname to scan, can be IP or hostname")
	portPtr := flag.String("ports", "80", "Ports to scan. Ex. 80 or 80,443,3306-3309")
	threadsPtr := flag.Int("threads", 100, "Number of threads to use, 100 is default and large numbers may create odd behavior")
	verbosePtr := flag.Bool("v", false, "Verbose output, including failures to connect")
	flag.Parse()
	hostname = *hostnamePtr
	ports = parsePorts(*portPtr)
	threads = *threadsPtr
	verbose = *verbosePtr
	return
}

func main() {
	// Adding concurrency
	passinhost, passinports, threads := parseCmdLineInput()

	// Creating a channel with a max of 100
	port_workers := make(chan int, threads)
	// Creating a channel for recording results and a Int slice for final
	results := make(chan int)
	var openports []int

	// The first set of workers going off, with a cap of 100
	for i := 0; i < cap(port_workers); i++ {
		go portScan(passinhost, port_workers, results)
	}

	// Sending the ports to the workers in the first set
	go func() {
		for _, ports := range passinports {
			port_workers <- ports
		}
	}()

	for range passinports {
		port := <-results
		if port != 0 {
			openports = append(openports, port)
			continue
		}
	}

	close(port_workers)
	close(results)
	sort.Ints(openports)
	for _, port := range openports {
		fmt.Printf("%d open\n", port)
	}
}

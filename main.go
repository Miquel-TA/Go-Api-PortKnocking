package main

import (
	"fmt"
	"net"
	"sync"
	"time"
)

// IPStatus holds the current knock step and the expiration time for an IP
type IPStatus struct {
	step       int
	expiration time.Time
}

var (
	knockSequence = []int{45010, 45030, 45020} // Sequence of ports to knock
	ipStatus      = sync.Map{}                 // Use sync.Map for thread-safe access
	expiredTime   = time.Time{}                // Default expiration time (a past time)
)

// Listen on all ports between 45000 and 45099 and handle connections
func monitorPorts() {
	for port := 45000; port <= 45099; port++ {
		go listenOnPort(port)
	}
}

// Listen for connections on a given port
func listenOnPort(port int) {
	address := fmt.Sprintf("0.0.0.0:%d", port)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		fmt.Printf("Error listening on port %d: %v\n", port, err)
		return
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("Error accepting connection on port %d: %v\n", port, err)
			continue
		}
		clientIP := extractIP(conn.RemoteAddr().String()) // Extract only the IP part
		conn.Close()                                      // Drop the connection to simulate a closed port

		handleKnock(clientIP, port)
	}
}

// Handle port knocking for an IP
func handleKnock(clientIP string, port int) {
	status := getIPStatus(clientIP)

	// Handle port knock sequence
	expectedPort := knockSequence[status.step]
	if port == expectedPort {
		// Correct port, progress the sequence
		fmt.Printf("Correct knock on port %d from %s, step %d\n", port, clientIP, status.step)
		status.step++

		// If the sequence is completed, whitelist the IP
		if status.step == len(knockSequence) {
			status.step = 0
			status.expiration = time.Now().Add(60 * time.Minute)
			fmt.Printf("IP %s whitelisted for 60 minutes\n", clientIP)
		}
	} else {
		// Incorrect port, reset the sequence
		fmt.Printf("Incorrect knock on port %d from %s, resetting sequence\n", port, clientIP)
		status.step = 0
		status.expiration = expiredTime
	}

	// Store the updated status as a value, not a pointer
	ipStatus.Store(clientIP, status)
}

// Get the status of an IP, defaulting to step 0 and expired time
func getIPStatus(clientIP string) IPStatus {
	statusInterface, exists := ipStatus.Load(clientIP)
	if !exists {
		// Default status: step 0, expired time
		return IPStatus{step: 0, expiration: expiredTime}
	}
	return statusInterface.(IPStatus)
}

// Start the API server on port 80
func startAPIServer() {
	listener, err := net.Listen("tcp", "0.0.0.0:80")
	if err != nil {
		fmt.Println("Error starting API server:", err)
		return
	}
	defer listener.Close()

	fmt.Println("API server started on port 80")
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting API connection:", err)
			continue
		}
		clientIP := extractIP(conn.RemoteAddr().String())

		status := getIPStatus(clientIP)

		if time.Now().Before(status.expiration) {
			fmt.Printf("Allowed access to API from %s\n", clientIP)
			go handleAPIRequest(conn)
		} else {
			fmt.Printf("Denied access to API from %s\n", clientIP)
			conn.Close()
		}
	}
}

// Handle the API request (valid HTTP response)
func handleAPIRequest(conn net.Conn) {
	defer conn.Close()

	// Define the HTTP response
	const response = "HTTP/1.1 200 OK\r\n" +
		"Content-Type: text/plain; charset=utf-8\r\n" +
		"Content-Length: 20\r\n" +
		"\r\n" +
		"Welcome to the API!\n"

	// Write the response
	_, err := conn.Write([]byte(response))
	if err != nil {
		fmt.Printf("Error sending response: %v\n", err)
	}
}


// Extract the IP part from a string like "IP:port"
func extractIP(addr string) string {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return addr // In case of error, return the original address
	}
	return host
}

// Start monitoring for knock sequences and API
func main() {
	fmt.Println("Monitoring ports for knocking sequence...")

	// Start monitoring ports for knocking
	monitorPorts()

	// Start the API server
	go startAPIServer()

	// Keep the program running
	select {}
}

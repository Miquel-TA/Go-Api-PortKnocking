package main

import (
	"fmt"
	"net"
	"sync"
	"time"
)

type ClientStatus struct {
	currentStep int
	whitelistExpiration time.Time
}

var (
	portKnockSequence    = []int{45010, 45030, 45020}
	clientStatuses       = sync.Map{}
	defaultExpiration    = time.Time{}
)

func monitorKnockPorts() {
	for port := 45000; port <= 45099; port++ {
		go listenOnPort(port)
	}
}

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
		clientIP := getClientIP(conn.RemoteAddr().String())
		conn.Close()

		processKnock(clientIP, port)
	}
}

func processKnock(clientIP string, port int) {
	clientStatus := getClientStatus(clientIP)

	expectedPort := portKnockSequence[clientStatus.currentStep]
	if port == expectedPort {
		// Correct knock!
		fmt.Printf("Successful knock on port %d from %s, step %d\n", port, clientIP, clientStatus.currentStep)
		clientStatus.currentStep++

		if clientStatus.currentStep == len(portKnockSequence) {
			clientStatus.currentStep = 0
			clientStatus.whitelistExpiration = time.Now().Add(60 * time.Minute)
			fmt.Printf("IP %s is whitelisted for 60 minutes\n", clientIP)
		}
	} else {
		// Incorrect knock!
		fmt.Printf("Failed knock on port %d from %s, resetting sequence\n", port, clientIP)
		clientStatus.currentStep = 0
		clientStatus.whitelistExpiration = defaultExpiration
	}

	clientStatuses.Store(clientIP, clientStatus)
}

func getClientStatus(clientIP string) ClientStatus {
	statusInterface, exists := clientStatuses.Load(clientIP)
	if !exists {
		return ClientStatus{currentStep: 0, whitelistExpiration: defaultExpiration}
	}
	return statusInterface.(ClientStatus)
}

func startHTTPServer() {
	listener, err := net.Listen("tcp", "0.0.0.0:80")
	if err != nil {
		fmt.Println("Error starting HTTP server:", err)
		return
	}
	defer listener.Close()

	fmt.Println("HTTP server started on port 80")
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting HTTP connection:", err)
			continue
		}
		clientIP := getClientIP(conn.RemoteAddr().String())

		clientStatus := getClientStatus(clientIP)
		if time.Now().Before(clientStatus.whitelistExpiration) {
			fmt.Printf("Granted access to HTTP server from %s\n", clientIP)
			go handleHTTPConnection(conn)
		} else {
			fmt.Printf("Denied access to HTTP server from %s\n", clientIP)
			conn.Close()
		}
	}
}

func handleHTTPConnection(conn net.Conn) {
	defer conn.Close()

	const httpResponse = "HTTP/1.1 200 OK\r\n" +
		"Content-Type: text/plain; charset=utf-8\r\n" +
		"Content-Length: 20\r\n" +
		"\r\n" +
		"Welcome to the API!\n"

	_, err := conn.Write([]byte(httpResponse))
	if err != nil {
		fmt.Printf("Error sending response: %v\n", err)
	}
}

func getClientIP(addr string) string {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return addr
	}
	return host
}

func main() {
	fmt.Println("Starting port monitoring for knock sequence...")

	// Start monitoring knock ports in the background.
	monitorKnockPorts()

	// Start the HTTP server in a separate goroutine.
	go startHTTPServer()

	// Block the main goroutine to keep the application running.
	select {}
}

// Variable Debug Web Server
//
// This utility implements an HTTP server that holds incoming requests and waits for manual
// user input (Enter key press) before sending responses. The server can accumulate multiple
// pending requests and release them all simultaneously with a single Enter press.
//
// This is useful for debugging scenarios where you need to:
//   - Test application behavior with delayed/slow HTTP responses
//   - Manually control response timing to simulate various network conditions
//   - Debug race conditions or timing-dependent issues
//   - Inspect application state while requests are in-flight
//   - Simulate thundering herd scenarios by releasing multiple requests at once
//   - Test concurrent request handling
//
// Behavior:
//   - Response headers (200 OK, Content-Type: application/json) are sent immediately
//   - Response body is held until Enter is pressed in the server terminal
//   - Each request is numbered and tracked
//   - A single Enter press releases ALL pending requests simultaneously
//   - Response body is JSON format: {"timestamp":"2025-12-15T12:34:56Z"}
//   - Timestamp is the current time in ISO-8601 format (UTC) when the response is sent
//
// Usage:
//   go run main.go              # Starts server on port 8080
//   PORT=3000 go run main.go    # Starts server on custom port

package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

type pendingRequest struct {
	requestTime  time.Time
	responseChan chan struct{}
	remoteAddr   string
	path         string
	method       string
}

type Server struct {
	mu              sync.Mutex
	pendingRequests []*pendingRequest
	requestCounter  int
}

func NewServer() *Server {
	return &Server{
		pendingRequests: make([]*pendingRequest, 0),
	}
}

func (s *Server) handleRequest(w http.ResponseWriter, r *http.Request) {
	requestTime := time.Now()

	// Create a pending request
	req := &pendingRequest{
		requestTime:  requestTime,
		responseChan: make(chan struct{}),
		remoteAddr:   r.RemoteAddr,
		path:         r.URL.Path,
		method:       r.Method,
	}

	// Add to pending requests
	s.mu.Lock()
	s.pendingRequests = append(s.pendingRequests, req)
	s.requestCounter++
	requestNum := s.requestCounter
	pendingCount := len(s.pendingRequests)
	s.mu.Unlock()

	fmt.Printf("\n[%s] Request #%d: %s %s from %s\n",
		requestTime.Format("15:04:05"), requestNum, r.Method, r.URL.Path, r.RemoteAddr)
	fmt.Printf("Pending requests: %d (Press ENTER to release all)\n", pendingCount)

	// Send response headers immediaapplication/jso
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)

	// Flush headers if possible
	if flusher, ok := w.(http.Flusher); ok {
		flusher.Flush()
	}

	// Wait for the signal to send response
	<-req.responseChan

	responseTime := time.Now()
	duration := responseTime.Sub(requestTime)

	fmt.Printf("[%s] Request #%d: Response body sent after waiting %s\n",
		responseTime.Format("15:04:05"), requestNum, duration)

	// Write the current timestamp in ISO-8601 format (UTC) as JSON
	timestamp := time.Now().UTC().Format(time.RFC3339)
	response := map[string]string{
		"timestamp": timestamp,
	}
	json.NewEncoder(w).Encode(response)
}

func (s *Server) waitForEnter() {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		// Get all pending requests
		s.mu.Lock()
		pendingRequests := s.pendingRequests
		count := len(pendingRequests)
		s.pendingRequests = make([]*pendingRequest, 0)
		s.mu.Unlock()

		if count == 0 {
			fmt.Println("No pending requests")
			continue
		}

		fmt.Printf("\nReleasing %d pending request(s)...\n", count)

		// Signal all pending requests to send their responses
		for _, req := range pendingRequests {
			close(req.responseChan)
		}
	}
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	server := NewServer()

	// Start the goroutine that waits for enter key
	go server.waitForEnter()

	http.HandleFunc("/", server.handleRequest)

	addr := fmt.Sprintf(":%s", port)
	fmt.Printf("Starting server on http://localhost%s\n", addr)
	fmt.Println("The server can hold multiple requests.")
	fmt.Println("Press ENTER to release ALL pending requests at once.")
	fmt.Println()

	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

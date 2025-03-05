package server

import (
	"fmt"
	"log"
	"net/http"
)

// StartHTTPServer starts an HTTP server on the given IP and port
func StartHTTPServer(ip string, port int) {
	address := fmt.Sprintf("%s:%d", ip, port)

	// Create a new ServeMux for this specific server
	mux := http.NewServeMux()

	// Register routes with the custom ServeMux, not the default global one
	mux.HandleFunc("/test", healthCheck)

	mux.HandleFunc("/", handleRequest)

	fmt.Printf("HTTP Server started on %v\n", address)

	// Start the HTTP server using the custom ServeMux
	err := http.ListenAndServe(address, mux)
	if err != nil {
		log.Fatalf("Error starting HTTP server on %v: %v", address, err)
	}
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	// Simple HTTP response for demonstration
	fmt.Fprintf(w, "Response from %s:%s", r.Host, r.URL.Port())
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	// Return 200 OK for the /test health check endpoint
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Healthy")
}

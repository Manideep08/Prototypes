package main

import (
	"fmt"
	"time"
	"net/http"
)

func main() {
	http.HandleFunc("/events", handleSSERequest)
	fmt.Printf("dude i am running on localhost")
	http.ListenAndServe(":8080", nil)
}

func handleSSERequest(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")


	for {
		currentTime := time.Now().Format("2006-01-02 15:04:05")

		fmt.Fprintf(w, "data: Current Time: %s\n\n", currentTime)

		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
			return
		}
		flusher.Flush()

		// Wait for 2 seconds before sending the next update
		time.Sleep(2 * time.Second)
	}
}
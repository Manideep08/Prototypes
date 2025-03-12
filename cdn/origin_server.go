package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

func main() {
	http.HandleFunc("/content", originHandler)
	fmt.Println("Origin Server running at :8080")
	http.ListenAndServe(":8080", nil) // Origin server listens on port 8080
}

func originHandler(w http.ResponseWriter, r *http.Request) {
	content, err := ioutil.ReadFile("content.txt") // Simulating a static file
	if err != nil {
		http.Error(w, "Content not found", http.StatusNotFound)
		return
	}

	fmt.Fprintf(w, string(content)) // Serve the content
}

package main

import "fmt"

func main() {
	server := NewRedisServer() // From entry.go
	fmt.Println("Redis prototype running on :6379 (Max Memory: 1MB)")
	runServer(server) // From server.go
}

package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"runtime"
	"strings"
	"time"
)

func runServer(rs *RedisServer) {
	runtime.GOMAXPROCS(1)
	listener, err := net.Listen("tcp", ":6379")
	if err != nil {
		fmt.Println("Failed to start server:", err)
		return
	}
	defer listener.Close()
	fmt.Println("Redis prototype running on :6379 (Max Memory: 1MB)")

	conns := make(map[net.Conn]bool)
	reader := make(map[net.Conn]*bufio.Reader)
	acceptChan := make(chan net.Conn, 10)

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				fmt.Println("Error accepting:", err)
				continue
			}
			acceptChan <- conn
		}
	}()

	for {
		// Accept new connections

		select {
		case conn := <-acceptChan:
			conns[conn] = true
			reader[conn] = bufio.NewReader(conn)
		default:
		}

		activeConns := make([]net.Conn, 0, len(conns))
		for conn := range conns {
			activeConns = append(activeConns, conn)
		}

		// Process connections
		for _, conn := range activeConns {
			if !conns[conn] { // Skip if deleted in previous iteration
				continue
			}
			fmt.Printf("Processing connection %v\n", conn.RemoteAddr())
			r := reader[conn]

			// Check connection state
			conn.SetReadDeadline(time.Now().Add(10 * time.Millisecond))
			_, err := r.Peek(1)
			if err != nil {
				if err == io.EOF || strings.Contains(err.Error(), "closed") || strings.Contains(err.Error(), "timeout") {
					fmt.Println("Connection dead or idle:", conn.RemoteAddr())
					conn.Close()
					delete(conns, conn)
					delete(reader, conn)
				} else {
					fmt.Println("Peek error:", err)
				}
				continue
			}

			// Read command
			conn.SetReadDeadline(time.Now().Add(50 * time.Millisecond))
			cmd, err := r.ReadString(' ')
			if err != nil {
				if strings.Contains(err.Error(), "timeout") {
					fmt.Println("Partial command, waiting:", conn.RemoteAddr())
					continue
				}
				fmt.Println("Read error:", conn.RemoteAddr(), err)
				conn.Close()
				delete(conns, conn)
				delete(reader, conn)
				continue
			}

			// fmt.Printf("Received command: %q\n", cmd)
			resp := rs.handleCommand(cmd)
			_, writeErr := conn.Write([]byte(resp))
			if writeErr != nil {
				fmt.Println("Write error:", conn.RemoteAddr(), writeErr)
				conn.Close()
				delete(conns, conn)
				delete(reader, conn)
				continue
			}

			// Check if client closed
			conn.SetReadDeadline(time.Now().Add(10 * time.Millisecond))
			if _, err := r.Peek(1); err == io.EOF {
				fmt.Println("Client closed after command:", conn.RemoteAddr())
				conn.Close()
				delete(conns, conn)
				delete(reader, conn)
			}
		}

		rs.cleanExpired(0)
		time.Sleep(10 * time.Millisecond)
	}
}

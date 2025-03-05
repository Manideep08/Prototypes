package main

import (
	"fmt"
	"sync"
	"time"

	"load_balancer/config"
	"load_balancer/loadbalancer"
	"load_balancer/server"
)

func main() {
	// Load server configurations
	servers := config.LoadConfig()

	// Start HTTP servers
	var wg sync.WaitGroup
	for _, srv := range servers {
		wg.Add(1)
		go func(srv config.ServerConfig) {
			defer wg.Done()
			fmt.Printf("server ip %v, port %v \n \n", srv.IP, srv.Port)
			go server.StartHTTPServer(srv.IP, srv.Port) // doing in sep routine because wg.Wait() wont let us go forward
		}(*srv)
	}

	wg.Wait()

	go loadbalancer.Start(servers)

	// Health check routine to periodically monitor the servers
	go func() {
		for {
			loadbalancer.HealthCheck(servers)
			time.Sleep(5 * time.Second)
		}
	}()

	// Keep the main routine alive
	// select {}
}

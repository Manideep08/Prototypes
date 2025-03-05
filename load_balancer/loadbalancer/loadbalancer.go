package loadbalancer

import (
	"fmt"
	"io"
	"load_balancer/config"
	"log"
	"net"
	"net/http"
	"sync"
)

type LoadBalancer struct {
	servers []*config.ServerConfig
	current int // for round robbin
	mu      sync.Mutex
	healthy map[string]bool
}

var lb *LoadBalancer

func Start(servers []*config.ServerConfig) {
	lb = &LoadBalancer{
		servers: servers,
		healthy: make(map[string]bool),
	}

	for _, s := range servers {
		lb.healthy[s.Address()] = true
	}

	fmt.Printf("starting loadbalancer dude ")
	lb.listenAndServe()

	// go
}

// HealthCheck periodically checks the health of servers
func HealthCheck(servers []*config.ServerConfig) {
	for _, srv := range servers {
		resp, err := http.Get(fmt.Sprintf("http://%s/test", srv.Address())) // Using /test for health check
		if err != nil || resp.StatusCode != http.StatusOK {
			log.Printf("Server %v is unhealthy", srv.Address())
			lb.mu.Lock()
			lb.healthy[srv.Address()] = false
			lb.mu.Unlock()
		} else {
			lb.mu.Lock()
			lb.healthy[srv.Address()] = true
			lb.mu.Unlock()
		}
	}
}

func (lb *LoadBalancer) listenAndServe() {
	listener, err := net.Listen("tcp", ":8000") // Load balancer listens on port 8000
	if err != nil {
		log.Fatalf("Error starting load balancer: %v", err)
	}

	log.Println(" ** Load balancer started on :8000\n\n")
	for {
		clientConn, err := listener.Accept()
		if err != nil {
			log.Println("Error accepting connection: ", err)
			continue
		}

		go lb.forwardRequest(clientConn)
	}
}

func (lb *LoadBalancer) getNextServer() *config.ServerConfig {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	for i := 0; i < len(lb.servers); i++ {
		lb.current = (lb.current + 1) % len(lb.servers)
		srv := lb.servers[lb.current]
		if lb.healthy[srv.Address()] {
			fmt.Printf("routing req to %v:%v", srv.IP, srv.Port)
			return srv
		}
	}
	return nil
}

func (lb *LoadBalancer) forwardRequest(clientConn net.Conn) {
	defer clientConn.Close()

	server := lb.getNextServer()
	if server == nil {
		log.Println("did not get any server")
		return
	}

	backendConn, err := net.Dial("tcp", server.Address())
	if err != nil {
		log.Printf("error while tcp dial to be server")
	}

	defer backendConn.Close()

	go func() {
		_, _ = io.Copy(clientConn, backendConn)
	}()

	_, _ = io.Copy(backendConn, clientConn)

}

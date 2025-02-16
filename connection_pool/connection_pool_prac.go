package main

import (
	"sync"
	"fmt"
)

type Connection struct {
	id int
}

// connection pool struct
type ConnectionPool struct {
	pool []*Connection
	mu sync.Mutex
	sm chan struct{}
	maxConnections int
}

func NewConnectionPool(maxConnections int) *ConnectionPool {
	cp := &ConnectionPool{
		pool: make([]*Connection, 0, maxConnections),
		sm: make(chan struct{}, maxConnections),
		maxConnections: maxConnections,
	}

	for j := 1; j <= maxConnections; j++ {
		conn := &Connection{id: j}
		cp.pool = append(cp.pool, conn)
	}


	return cp
}

func GetNewConnection(cp *ConnectionPool) *Connection {
	cp.sm <- struct{}{}

	if len(cp.pool) == 0 {
		fmt.Printf("Bro pool size is zerooooo\n")
		return nil
	}

	cp.mu.Lock()
	defer cp.mu.Unlock()
	conn := cp.pool[0]
	cp.pool = cp.pool[1:]
	fmt.Printf("Giving a connection bro %v\n", conn.id)
	return conn

}

func ReleaseConncetion(conn *Connection, cp *ConnectionPool) {
	cp.mu.Lock()
	defer cp.mu.Unlock()
	cp.pool = append(cp.pool, conn)
	fmt.Printf("Releasing connection bro %d\n ", conn.id)
	<-cp.sm
}


func PerformWork(conn *Connection, worker int, cp *ConnectionPool) {
	fmt.Printf("Manideep requesting worker %d conn %d\n", worker, conn.id)
	ReleaseConncetion(conn, cp)
}


func main() {
	pools := NewConnectionPool(3)

	fmt.Printf("Initialising pools bro")
	// fmt.Printf(pools)

	var wg sync.WaitGroup

	for i := 1 ; i <= 10 ; i++ {
		wg.Add(1)
		go func(workerid int) {
			defer wg.Done()
			conn := GetNewConnection(pools)
			PerformWork(conn, workerid, pools)
		}(i)

	}

	wg.Wait()
}


package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"time"
)

var (
	cacheItems     = make(map[string]CacheItem)
	cacheMutex     = sync.RWMutex{}
	requestCounter = make(map[string]int)
)

type CacheItem struct {
	content       []byte
	expiresAt     time.Time
	lastUpdatedAt time.Time
}

const CacheExpirationDuration = time.Minute * 5
const MaxRequestsPerMinute = 10 // Rate limit for each IP

func main() {

	go resetRequestCounter()

	http.HandleFunc("/content", ServeContent)
	fmt.Println("CDN Edge Server running at :8081")
	http.ListenAndServe(":8081", nil)
}

func ServeContent(w http.ResponseWriter, r *http.Request) {
	clientIp := r.RemoteAddr

	if !isAllowed(clientIp) {
		http.Error(w, "Too many requests! Slow down, chef!", http.StatusTooManyRequests)
		return
	}

	cacheMutex.RLock()
	cacheItem, found := cacheItems[r.URL.Path]
	cacheMutex.RUnlock()

	// check to avoid serving stale content
	if found && time.Now().Before(cacheItem.expiresAt) {
		fmt.Println("Serving from cache:", r.URL.Path)
		w.Write(cacheItem.content)
		return
	}

	fmt.Println("Cache miss. Fetching from origin for:", r.URL.Path)

	content, err := FetchContentFromOrigin("http://localhost:8080" + r.URL.Path)
	if err != nil {
		http.Error(w, "Oops! Couldn't fetch content from the origin. The kitchen might be down!", http.StatusBadGateway)
		return
	}

	fmt.Printf("Cache set for %s, will expire at %v\n", r.URL.Path, time.Now().Add(CacheExpirationDuration))

	cacheMutex.Lock()
	cacheItems[r.URL.Path] = CacheItem{
		content:       content,
		expiresAt:     time.Now().Add(CacheExpirationDuration),
		lastUpdatedAt: time.Now(),
	}
	cacheMutex.Unlock()

	w.Write(content)

}

func FetchContentFromOrigin(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return content, nil
}

// isAllowed checks whether the client IP has exceeded the request limit
func isAllowed(clientIP string) bool {
	requestCounter[clientIP]++
	return requestCounter[clientIP] <= MaxRequestsPerMinute
}

func resetRequestCounter() {
	for {
		time.Sleep(time.Minute)
		cacheMutex.Lock()
		requestCounter = make(map[string]int) // Reset request count every minute
		cacheMutex.Unlock()
	}
}

package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"hash/crc32"
	"log"
	"net/http"
	"sync"
	"time"

	_ "github.com/lib/pq"
)

var dbConnections = map[int]*sql.DB{}
var dbMutex sync.Mutex

func getDbConnection(shardId int) *sql.DB {
	dbMutex.Lock()
	defer dbMutex.Unlock()

	if dbConnections[shardId] == nil {
		var connStr string
		switch shardId {
		case 0:
			connStr = "user=allenkimanideep dbname=key_value_store host=localhost port=5432 sslmode=disable"
		case 1:
			connStr = "user=allenkimanideep dbname=key_value_store host=localhost port=5433 sslmode=disable"
		case 2:
			connStr = "user=allenkimanideep dbname=key_value_store host=localhost port=5434 sslmode=disable"
		}

		db, err := sql.Open("postgres", connStr)
		if err != nil {
			log.Fatal(err)
		}
		dbConnections[shardId] = db
	}
	return dbConnections[shardId]
}

func get(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	shardId := shardKey(key)

	db := getDbConnection(shardId)

	var value string

	err := db.QueryRow("SELECT value FROM key_value_store where key = $1 AND (expires_at IS NULL OR expires_at > CURRENT_TIMESTAMP) AND deleted_at is NULL", key).Scan(&value)
	if err != nil {
		fmt.Printf("error while fetching key %v\n", err)
		http.Error(w, "key not found dude", http.StatusNotFound)
		return
	}

	fmt.Fprintf(w, "Key: %s Value: %s", key, value)

}

func deleteKey(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	shardId := shardKey(key)

	db := getDbConnection(shardId)

	_, err := db.Exec("UPDATE key_value_store SET deleted_at = CURRENT_TIMESTAMP where key = $1", key)
	if err != nil {
		fmt.Printf("error while deleting key dude %v\n", err)
		http.Error(w, "unable to delete key dude", http.StatusInternalServerError)
	}

	fmt.Fprintf(w, "Successfully deleted Key %s", key)
}

func put(w http.ResponseWriter, r *http.Request) {

	var input struct {
		Key       string `json:"key"`
		Value     string `json:"value"`
		ExpiresIn int64  `json:"expires_in"`
	}

	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	if input.Key == "" || input.Value == "" {
		http.Error(w, "Key or value cannot be empty", http.StatusBadRequest)
		return
	}

	expireAt := time.Now().Add(time.Duration(input.ExpiresIn) * time.Second)

	shardId := shardKey(input.Key)
	db := getDbConnection(shardId)

	_, err = db.Exec("INSERT INTO key_value_store (key, value, expires_at) VALUES ($1, $2, $3) ON CONFLICT (key) DO UPDATE SET value = $2, expires_at = $3", input.Key, input.Value, expireAt)
	if err != nil {
		fmt.Errorf("dude error while inserting in DB %v", err)
		http.Error(w, "Error while inserting or updating key", http.StatusInternalServerError)
	}

	fmt.Fprintf(w, "Done with operation boss for key %s", input.Key)

}

func shardKey(key string) int {
	hash := crc32.ChecksumIEEE([]byte(key))
	return int(hash % 3)
}

func expireKeys() {
	time.Sleep(time.Duration(5) * time.Minute)

	for i := 1; i <= 3; i++ {
		db := getDbConnection(i)
		_, err := db.Exec("UPDATE key_value_store SET deleted_at = CURRENT_TIMESTAMP where expires_at < CURRENT_TIMESTAMP AND deleted_at IS NULL")
		if err != nil {
			log.Printf("error while expiring keys on shard %d: %v", i, err)
		}
	}
}

func main() {
	// triggering a go routine to expire all keys
	go expireKeys()

	http.HandleFunc("/put", put)
	http.HandleFunc("/get", get)
	http.HandleFunc("/delete", deleteKey)

	http.ListenAndServe(":8080", nil)
}

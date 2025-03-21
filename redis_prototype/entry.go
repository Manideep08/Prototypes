package main

import (
	"fmt"
	"math/rand"
	"time"
	"unsafe"
)

const MaxMemory = 500 * 1024 * 1024 // 1MB memory limit

type Entry struct {
	Key       string
	Value     string
	Expiry    time.Time
	KeySize   int
	ValueSize int
}

type RedisServer struct {
	store       map[string]*Entry
	totalMemory int
}

func NewRedisServer() *RedisServer {
	return &RedisServer{
		store:       make(map[string]*Entry),
		totalMemory: 0,
	}
}

func (rs *RedisServer) cleanExpired(targetSpace int) bool {
	const sampleSize = 20
	const targetDeletes = 5
	const maxCycles = 10

	now := time.Now()
	deleted := 0
	cycles := 0
	freedSpace := 0
	keysChecked := 0

	for cycles < maxCycles && (deleted < targetDeletes || freedSpace < targetSpace) && len(rs.store) > 0 {
		totalKeys := len(rs.store) // Live count
		if totalKeys == 0 {
			fmt.Println("No keys to clean")
			return true
		}

		offset := rand.Intn(totalKeys)
		checkedThisCycle := 0
		anyExpired := false

		for key, entry := range rs.store {
			if keysChecked < offset {
				keysChecked++
				continue
			}
			if checkedThisCycle >= sampleSize {
				break
			}

			if !entry.Expiry.IsZero() && now.After(entry.Expiry) {
				entrySize := int(unsafe.Sizeof(*entry)) + entry.KeySize + entry.ValueSize
				rs.totalMemory -= entrySize
				freedSpace += entrySize
				delete(rs.store, key)
				fmt.Printf("Expired key '%s', freed %d, total memory: %d\n", key, entrySize, rs.totalMemory)
				deleted++
				anyExpired = true
			}
			checkedThisCycle++
			keysChecked++
		}

		cycles++
		if !anyExpired && checkedThisCycle > 0 { // No expirations found in this cycle
			fmt.Println("No expired keys found, exiting early")
			break
		}

		if keysChecked >= totalKeys {
			keysChecked = 0
		}
	}
	fmt.Printf("Clean cycle done: deleted %d, freed %d\n", deleted, freedSpace)
	return freedSpace >= targetSpace
}

func (rs *RedisServer) canAddEntry(key, value string) bool {
	keyMem := int(unsafe.Sizeof(key)) + len(key)
	valMem := int(unsafe.Sizeof(value)) + len(value)
	entryMem := int(unsafe.Sizeof(Entry{})) + keyMem + valMem

	if rs.totalMemory+entryMem <= MaxMemory {
		return true
	}

	delta := (rs.totalMemory + entryMem) - MaxMemory
	return rs.cleanExpired(delta)
}

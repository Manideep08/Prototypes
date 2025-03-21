package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"
	"unsafe"
)

func (rs *RedisServer) set(key, value string) string {
	// Check if we can add this entry
	if !rs.canAddEntry(key, value) {
		return "-ERR out of memory\r\n"
	}

	// Remove old entry if exists
	if oldEntry, exists := rs.store[key]; exists {
		entrySize := int(unsafe.Sizeof(*oldEntry)) + oldEntry.KeySize + oldEntry.ValueSize
		rs.totalMemory -= entrySize
	}

	// Calculate real memory sizes
	keyMem := int(unsafe.Sizeof(key)) + len(key)
	valMem := int(unsafe.Sizeof(value)) + len(value)
	entry := &Entry{
		Key:       key,
		Value:     value,
		KeySize:   keyMem,
		ValueSize: valMem,
	}
	entrySize := int(unsafe.Sizeof(*entry)) + keyMem + valMem
	rs.totalMemory += entrySize
	rs.store[key] = entry
	fmt.Printf("SET %s: Added %d bytes, total: %d\n", key, entrySize, rs.totalMemory)
	return "+OK\r\n"
}

func (rs *RedisServer) get(key string) string {
	entry, exists := rs.store[key]
	if !exists || (!entry.Expiry.IsZero() && time.Now().After(entry.Expiry)) {
		return "$-1\r\n"
	}
	return fmt.Sprintf("$%d\r\n%s\r\n", len(entry.Value), entry.Value)
}

func (rs *RedisServer) expire(key string, seconds int) string {
	entry, exists := rs.store[key]
	if !exists {
		return ":0\r\n"
	}
	entry.Expiry = time.Now().Add(time.Duration(seconds) * time.Second)
	return ":1\r\n"
}

func (rs *RedisServer) ttl(key string) string {
	entry, exists := rs.store[key]
	if !exists {
		return ":-2\r\n"
	}
	if entry.Expiry.IsZero() {
		return ":-1\r\n"
	}
	if time.Now().After(entry.Expiry) {
		return ":-2\r\n"
	}
	ttl := int(entry.Expiry.Sub(time.Now()).Seconds())
	fmt.Printf("TTL calculated: %d\n", ttl) // Debug
	return fmt.Sprintf(":%d\r\n", ttl)
}

func (rs *RedisServer) del(key string) string {
	entry, exists := rs.store[key]
	if !exists || (!entry.Expiry.IsZero() && time.Now().After(entry.Expiry)) {
		return ":0\r\n" // Key not found or already expired
	}
	entrySize := int(unsafe.Sizeof(*entry)) + entry.KeySize + entry.ValueSize
	rs.totalMemory -= entrySize
	delete(rs.store, key)
	fmt.Printf("DEL %s: Freed %d bytes, total: %d\n", key, entrySize, rs.totalMemory)
	return ":1\r\n"
}

/*
command explanation

*3\r\n$3\r\nSET\r\n$4\r\nkey1\r\n$6\r\nvalue1\r\n
*3: Array with 3 items.
$3\r\nSET: Command “SET” (3 bytes).
$4\r\nkey1: Key “key1” (4 bytes).
$6\r\nvalue1: Value “value1” (6 bytes).
Meaning: SET key1 value1.
*/

func (rs *RedisServer) handleCommand(cmd string) string {
	fmt.Printf("Received command: %q\n", cmd)
	parts := strings.Split(cmd, "\r\n")
	fmt.Printf("Parts: %v\n", parts)
	if len(parts) < 3 || parts[0][0] != '*' {
		return "-ERR invalid command\r\n"
	}
	numArgs, _ := strconv.Atoi(parts[0][1:])
	if len(parts)-1 < 2*numArgs {
		return "-ERR malformed command\r\n"
	}
	command := strings.ToUpper(parts[2])
	switch command {
	case "SET":
		if numArgs != 3 {
			return "-ERR wrong number of arguments for SET\r\n"
		}
		return rs.set(parts[4], parts[6])
	case "GET":
		if numArgs != 2 {
			return "-ERR wrong number of arguments for GET\r\n"
		}
		return rs.get(parts[4])
	case "EXPIRE":
		if numArgs != 3 {
			return "-ERR wrong number of arguments for EXPIRE\r\n"
		}
		seconds, _ := strconv.Atoi(parts[6])
		return rs.expire(parts[4], seconds)
	case "TTL":
		if numArgs != 2 {
			return "-ERR wrong number of arguments for TTL\r\n"
		}
		return rs.ttl(parts[4])
	case "DEL":
		if numArgs != 2 {
			return "-ERR wrong number of arguments for DEL\r\n"
		}
		return rs.del(parts[4])
	default:
		return fmt.Sprintf("-ERR unknown command '%s'\r\n", command)
	}
}

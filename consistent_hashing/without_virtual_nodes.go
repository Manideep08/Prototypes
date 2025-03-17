package main

import (
	"crypto/sha256"
	"fmt"
	"sort"
)

func hash(key string) uint32 {
	h := sha256.New()
	h.Write([]byte(key))
	hashBytes := h.Sum(nil)
	return uint32(hashBytes[0])<<24 | uint32(hashBytes[1])<<16 | uint32(hashBytes[2])<<8 | uint32(hashBytes[3])

}

type ConsistentHash struct {
	nodes   []uint32
	nodeMap map[uint32]string
}

func NewConsistentHash() *ConsistentHash {
	return &ConsistentHash{
		nodes:   []uint32{},              // empty slice of nodes
		nodeMap: make(map[uint32]string), // empty map for node hashes to names
	}
}

func (c *ConsistentHash) AddNode(node string) {
	nodeHash := hash(node)
	c.nodes = append(c.nodes, nodeHash)
	c.nodeMap[nodeHash] = node

	sort.Slice(c.nodes, func(i int, j int) bool { return c.nodes[i] < c.nodes[j] })
}

func (c *ConsistentHash) GetNode(key string) string {
	keyHash := hash(key)

	idx := sort.Search(len(c.nodes), func(i int) bool { return c.nodes[i] >= keyHash })

	if idx == len(c.nodes) {
		idx = 0
	}

	return c.nodeMap[c.nodes[idx]]
}

func (c *ConsistentHash) PrintNodes() {
	fmt.Println("Nodes in the ring:")
	for _, nodeHash := range c.nodes {
		fmt.Printf("Node: %s, Hash: %d\n", c.nodeMap[nodeHash], nodeHash)
	}
}

func main() {
	// Create the consistent hash ring
	ch := NewConsistentHash()

	// Adding finite nodes (servers)
	ch.AddNode("ServerA")
	ch.AddNode("ServerB")
	ch.AddNode("ServerC")
	ch.AddNode("ServerD")

	// Print the current state of the hash ring (the nodes and their hashed values)
	ch.PrintNodes()

	// // Assigning files (keys) to servers (nodes) and printing the result
	// fmt.Println("\nFile 'file1.txt' goes to:", ch.GetNode("file1.txt"))
	// fmt.Println("File 'file2.txt' goes to:", ch.GetNode("file2.txt"))
	// fmt.Println("File 'file3.txt' goes to:", ch.GetNode("file3.txt"))

	hotKeys := []string{
		"hotfile1", "hotfile2", "hotfile3", "hotfile4", "hotfile5", "hotfile6", "hotfile7",
		"hotfile8", "hotfile9", "hotfile10", "hotfile11", "hotfile12", "hotfile13", "hotfile14",
		"hotfile15", "hotfile16", "hotfile17", "hotfile18", "hotfile19", "hotfile20",
	} // Severe skew towards specific hot keys

	coldKeys := []string{
		"file1", "file2", "file3", "file4", "file5",
	} // Fewer cold keys for more balance

	// Assigning files (keys) to servers (nodes)
	fmt.Println("\nAssigning hot keys:")
	for _, key := range hotKeys {
		fmt.Printf("File '%s' goes to: %s\n", key, ch.GetNode(key))
	}

	fmt.Println("\nAssigning cold keys:")
	for _, key := range coldKeys {
		fmt.Printf("File '%s' goes to: %s\n", key, ch.GetNode(key))
	}

}

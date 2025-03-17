package main

import (
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"math/big"
	"sort"
	"strconv"
)

// Hash function: simple SHA-256-based hash
func hash(s string) uint32 {
	h := sha256.New()
	h.Write([]byte(s))
	hashBytes := h.Sum(nil)
	return uint32(hashBytes[0])<<24 | uint32(hashBytes[1])<<16 | uint32(hashBytes[2])<<8 | uint32(hashBytes[3])
}

// Consistent Hashing with Virtual Nodes struct
type ConsistentHash struct {
	nodes            []uint32          // Sorted hash values (including virtual nodes)
	nodeMap          map[uint32]string // Maps hash to node names (virtual node positions)
	virtualNodeCount int               // Number of virtual nodes per real node
}

// NewConsistentHash creates a new consistent hashing object with virtual nodes
func NewConsistentHash(virtualNodeCount int) *ConsistentHash {
	return &ConsistentHash{
		nodes:            []uint32{},              // empty slice of nodes
		nodeMap:          make(map[uint32]string), // empty map for virtual nodes
		virtualNodeCount: virtualNodeCount,        // number of virtual nodes per real node
	}
}

// AddNode adds a server (node) to the ring with multiple virtual nodes
func (ch *ConsistentHash) AddNode(node string) {
	// Add the real node itself
	nodeHash := hash(node) // hash the real node
	ch.nodes = append(ch.nodes, nodeHash)
	ch.nodeMap[nodeHash] = node

	// For each real node, create virtual nodes (multiple hash positions)
	for i := 0; i < ch.virtualNodeCount; i++ {
		// Create a unique virtual node name for each virtual node (e.g., ServerA#1, ServerA#2)
		// virtualNodeName := node + "#" + strconv.Itoa(i)
		// virtualNodeHash := hash(virtualNodeName) // hash the virtual node name

		// adding below to spread virtual nodes spread across above ring as above approach did not work
		randNum, err := rand.Int(rand.Reader, big.NewInt(10000))
		if err != nil {
			panic(err) // Simple error handling for demo
		}
		// Randomize the virtual node name, e.g., "ServerA#2-7341"
		virtualNodeName := node + "#" + strconv.Itoa(i) + "-" + randNum.String()
		virtualNodeHash := hash(virtualNodeName) // hash the virtual node name

		// Add the virtual node hash to the ring
		ch.nodes = append(ch.nodes, virtualNodeHash)

		// Map the virtual node hash to the real node name
		ch.nodeMap[virtualNodeHash] = node
	}

	// Sort the virtual and real nodes in ascending order to maintain the ring structure
	sort.Slice(ch.nodes, func(i, j int) bool { return ch.nodes[i] < ch.nodes[j] })
}

// GetNode returns the server responsible for a given key
func (ch *ConsistentHash) GetNode(key string) string {
	keyHash := hash(key) // hash the key (file name)

	// Find the first virtual node whose hash is greater than or equal to the key's hash
	idx := sort.Search(len(ch.nodes), func(i int) bool { return ch.nodes[i] >= keyHash })

	// If keyHash is greater than the largest virtual node hash, wrap around to the first virtual node
	if idx == len(ch.nodes) {
		idx = 0
	}

	// Return the real node that corresponds to the virtual node
	return ch.nodeMap[ch.nodes[idx]]
}

// PrintNodes prints the current list of virtual nodes and their corresponding real nodes
func (ch *ConsistentHash) PrintNodes() {
	fmt.Println("Virtual and Real Nodes in the ring:")
	for _, nodeHash := range ch.nodes {
		virtualNodeName := ch.nodeMap[nodeHash]
		fmt.Printf("Node: %s, Hash: %d\n", virtualNodeName, nodeHash)
	}
}

func main() {
	// Create the consistent hash ring with 3 virtual nodes per real node
	ch := NewConsistentHash(4)

	// Adding nodes (servers) to the ring
	ch.AddNode("ServerA")
	ch.AddNode("ServerB")
	ch.AddNode("ServerC")
	ch.AddNode("ServerD")

	// Print nodes to see the distribution of hash values
	ch.PrintNodes()

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

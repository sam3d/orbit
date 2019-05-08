package engine

import (
	"encoding/hex"
	"math/rand"
	"net"
)

// Node is a server attached to the system in some capacity.
type Node struct {
	ID      string `json:"id"`      // The unique ID of the node
	Address net.IP `json:"address"` // The address of the node

	RPCPort     int `json:"rpc_port"`
	RaftPort    int `json:"raft_port"`
	SerfPort    int `json:"serf_port"`
	WANSerfPort int `json:"wan_serf_port"`

	SwapSize   uint `json:"swap_size"`  // The size of the swap in MB
	Swappiness uint `json:"swappiness"` // Likelihood of swapping (0 - 100)
}

// Nodes is a list of nodes.
type Nodes []Node

// GenerateNodeID generate an ID for a new node. After generating, it will
// search the existing list of nodes and ensure that it is unique.
func (n *Nodes) GenerateNodeID() string {
search:
	for {
		// Generate random 32 byte slice.
		b := make([]byte, 32)
		rand.Read(b)
		id := hex.EncodeToString(b)

		// Search for duplicates, and if it matches, reset the search.
		for _, node := range *n {
			if id == node.ID {
				continue search
			}
		}

		// If we made it this far, there were no duplicates.
		return id
	}
}

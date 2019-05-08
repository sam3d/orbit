package engine

import (
	"encoding/hex"
	"math/rand"
	"net"
)

// NodeRole is a the type of role that the node can be.
type NodeRole string

const (
	// RoleManager means that the role is a manager node for the cluster.
	RoleManager NodeRole = "MANAGER"
	// RoleWorker means that the node participates in the gossip of the cluster,
	// but importantly not the operation and organisation of it.
	RoleWorker = "WORKER"

	// RoleLoadBalancer means that the node serves as an ingress point for the
	// cluster.
	RoleLoadBalancer = "LOAD_BALANCER"
	// RoleStorage means that the node is responsible for the storage of the
	// general contents of the cluster.
	RoleStorage = "STORAGE"
	// RoleBuilder designates a node as an image builder for the cluster. This is
	// where the repo contents will end up for storage.
	RoleBuilder = "BUILDER"
)

// Node is a server attached to the system in some capacity.
type Node struct {
	ID      string `json:"id"`      // The unique ID of the node
	Address net.IP `json:"address"` // The address of the node

	RPCPort     int `json:"rpc_port"`
	RaftPort    int `json:"raft_port"`
	SerfPort    int `json:"serf_port"`
	WANSerfPort int `json:"wan_serf_port"`

	Roles      []NodeRole `json:"node_roles"` // What roles this node serves
	SwapSize   uint       `json:"swap_size"`  // The size of the swap in MB
	Swappiness uint       `json:"swappiness"` // Likelihood of swapping (0 - 100)
}

// HasRole returns whether or not a node has a given role.
func (n Node) HasRole(role NodeRole) bool {
	for _, r := range n.Roles {
		if r == role {
			return true
		}
	}
	return false
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

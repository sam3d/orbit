package engine

import (
	"encoding/hex"
	"math/rand"
)

// StoreState is the all-encompassing state of the cluster. The operations are
// performed to this after being cast to a finite state machine, and otherwise
// won't be able to make any changes.
//
// Important to note is that the state is not aware of its distributed nature,
// and is simply for keeping track of the current data.
type StoreState struct {
	Namespaces   Namespaces   `json:"namespaces"`
	Users        Users        `json:"users"`
	Nodes        Nodes        `json:"nodes"`
	Routers      Routers      `json:"routers"`
	Certificates Certificates `json:"certificates"`
	Volumes      Volumes      `json:"volumes"`

	ManagerJoinToken string `json:"manager_join_token"`
	WorkerJoinToken  string `json:"worker_join_token"`
}

// Namespace is a location where certain elements exist in. The elements in
// question still need to be globally unique, however they can be defined within
// existing locations.
type Namespace struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Namespaces is a list of namespaces.
type Namespaces []Namespace

// GenerateID will generate a unique ID for a namespace.
func (n *Namespaces) GenerateID() string {
search:
	for {
		b := make([]byte, 8)
		rand.Read(b)
		id := hex.EncodeToString(b)

		for _, namespace := range *n {
			if namespace.ID == id {
				continue search
			}
		}

		return id
	}
}

// Find will find a namespace with the given name or identifier and return that
// namespace. Will return nil if it could not be found.
func (n *Namespaces) Find(id string) *Namespace {
	for _, namespace := range *n {
		if namespace.ID == id || namespace.Name == id {
			return &namespace
		}
	}
	return nil
}

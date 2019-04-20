package engine

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

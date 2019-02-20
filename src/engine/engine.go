// Package engine provides the all-encompassing interface to the Orbit
// background operations. This includes replicated state management, gossip
// control, and ensuring that the state is maintained for the respective nodes.
package engine

import "orbit.sh/engine/api"

// Engine is the primary all-encompassing struct for the primary Orbit
// operations. This means that all of the top-level features such as the
// replicated state store and REST API are located here.
type Engine struct {
	APIServer *api.Server
}

// New creates a new instance of the engine.
func New() *Engine {
	return &Engine{
		APIServer: api.NewServer("/var/run/orbit.sock", 6501),
	}
}

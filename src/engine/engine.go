// Package engine provides the all-encompassing interface to the Orbit
// background operations. This includes replicated state management, gossip
// control, and ensuring that the state is maintained for the respective nodes.
package engine

import (
	"log"

	"orbit.sh/engine/api"
	"orbit.sh/engine/store"
)

// Engine is the primary all-encompassing struct for the primary Orbit
// operations. This means that all of the top-level features such as the
// replicated state store and REST API are located here.
type Engine struct {
	API   *api.Server
	Store *store.Store

	Logger *log.Logger
}

// New creates a new instance of the engine.
func New() *Engine {
	return &Engine{
		API:   api.New(),
		Store: store.New(),
	}
}

// Start starts the engine and all of its subcomponents. This is dependent on
// state, so for example if the cluster still has yet to be set up, then it
// won't start the store.
func (e *Engine) Start() error {
	return nil
}

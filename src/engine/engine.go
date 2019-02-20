// Package engine provides the all-encompassing interface to the Orbit
// background operations. This includes replicated state management, gossip
// control, and ensuring that the state is maintained for the respective nodes.
package engine

import (
	"log"

	"orbit.sh/engine/api"
	"orbit.sh/engine/store"
)

var (
	// APIServer is the main instance of the API server running.
	APIServer *api.Server
	// Store is the main instance of the replicated store.
	Store *store.Store
	// Started is whether or not the engine is started.
	Started = false
	// Logger is the main logging instance for the engine.
	Logger *log.Logger
)

// Start will start the engine.
func Start() {
	if Started {
		Logger.Println("already started, ignoring request to start")
		return
	}

	go startAPIServer()

	Started = true
}

// startAPIServer starts a new instance of the API server.
func startAPIServer() {
	// Create the API server.
	srv, err := api.New()
	if err != nil {
		panic(err)
	}
	APIServer = srv

	// Start the API server.
	err = srv.Start()
	panic(err)
}

// Stop will halt the operation of the engine.
func Stop() {
	if !Started {
		Logger.Println("already stopped, ignoring request to stop")
		return
	}
	Started = false
}

package engine

import (
	"orbit.sh/engine/api"
)

var (
	// APIServer is the main instance of the API server running.
	APIServer *api.Server
)

// Start will start the engine.
func Start() {
	go startAPIServer()
	select {} // Pause
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

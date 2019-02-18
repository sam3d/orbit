package engine

import (
	"orbit.sh/engine/api"
)

// Start will start the engine
func Start() {
	// Start the API server
	apiSrv := api.New()
	apiSrv.SocketPath = "" // Disable socket listener
	err := apiSrv.Start()
	panic(err)
}

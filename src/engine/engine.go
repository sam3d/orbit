package engine

import (
	"orbit.sh/engine/api"
)

// Start will start the engine
func Start() {
	// Start the API server
	apiSrv, err := api.New()
	apiSrv.SocketPath = ""
	if err != nil {
		panic(err)
	}
	err = apiSrv.Start()
	panic(err)
}

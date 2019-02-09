package engine

import (
	"fmt"
	"log"
)

// Started is whether or not the engine process is running.
var Started = false

// Start will start the engine.
func Start() error {
	if Started {
		return fmt.Errorf("Engine has already been started")
	}

	log.Println("[INFO] engine: Starting the engine...")
	defer log.Println("[INFO] engine: Started")
	Started = true

	return nil
}

// Stop will stop the engine.
func Stop() error {
	if !Started {
		return fmt.Errorf("Engine has already been stopped")
	}

	log.Println("[INFO] engine: Stopping the engine...")
	defer log.Println("[INFO] engine: Stopped")
	Started = false

	return nil
}

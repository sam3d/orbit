// Package engine provides the all-encompassing interface to the Orbit
// background operations. This includes replicated state management, gossip
// control, and ensuring that the state is maintained for the respective nodes.
package engine

import (
	"log"
)

// Engine is the primary all-encompassing struct for the primary Orbit
// operations. This means that all of the top-level features such as the
// replicated state store and REST API are located here.
type Engine struct {
	APIServer *APIServer
	Store     *Store

	Status     Status
	DataPath   string
	ConfigFile string
}

// New creates a new instance of the engine.
func New() *Engine {
	e := &Engine{
		Status:     Init,
		DataPath:   "/var/orbit",
		ConfigFile: "config.json",
	}

	e.Store = NewStore(e)
	e.APIServer = NewAPIServer(e)

	return e
}

// Status is an enum about the current state of the engine.
type Status uint8

const (
	// Init is the first opening state of the engine and means that the config has
	// not yet been loaded.
	Init Status = iota
	// Setup is when the engine has not yet been bootstrapped.
	Setup
	// Ready is when the engine can properly be used.
	Ready
)

func (s Status) String() string {
	switch s {
	case Init:
		return "init"
	case Setup:
		return "setup"
	case Ready:
		return "ready"
	default:
		return ""
	}
}

// Start starts the engine and all of its subcomponents. This is dependent on
// state, so for example if the cluster still has yet to be set up, then it
// won't start the store.
func (e *Engine) Start() error {
	log.Println("[INFO] engine: Starting...")

	errCh := make(chan error) // Main error channel closure

	// Read in the config
	if err := e.readConfig(); err != nil {
		return err
	}

	// Start the API Server.
	go func() {
		errCh <- e.APIServer.Start()
	}()

	// Open the Store.
	go func() {
		if e.Status == Ready && e.Store.AdvertiseAddr != nil {
			errCh <- e.Store.Open()
		}
	}()

	// Monitor started progress on each component.
	go func() {
		<-e.APIServer.Started()
		<-e.Store.Started()
		log.Println("[INFO] engine: Started")
	}()

	return <-errCh
}

// Stop will stop the operation of the engine instance.
func (e *Engine) Stop() error {
	log.Println("[INFO] engine: Stopping...")
	log.Println("[INFO] engine: Stopped")
	return nil
}

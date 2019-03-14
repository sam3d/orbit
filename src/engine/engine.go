// Package engine provides the all-encompassing interface to the Orbit
// background operations. This includes replicated state management, gossip
// control, and ensuring that the state is maintained for the respective nodes.
package engine

import (
	"log"
	"os"
	"path/filepath"
)

// Engine is the primary all-encompassing struct for the primary Orbit
// operations. This means that all of the top-level features such as the
// replicated state store and REST API are located here.
type Engine struct {
	APIServer *APIServer
	RPCServer *RPCServer
	Store     *Store

	Status     Status
	DataPath   string
	ConfigFile string
}

// New creates a new instance of the engine.
func New() *Engine {
	e := &Engine{
		Status:     StatusInit,
		DataPath:   "/var/orbit",
		ConfigFile: "config.json",
	}

	e.Store = NewStore(e)
	e.APIServer = NewAPIServer(e)
	e.RPCServer = NewRPCServer(e)

	return e
}

// Status is an enum about the current state of the engine.
type Status uint8

const (
	// StatusInit is the first opening state of the engine and means that the config has
	// not yet been loaded.
	StatusInit Status = iota
	// StatusSetup is when the engine has not yet been bootstrapped.
	StatusSetup
	// StatusReady is when the engine can properly be used.
	StatusReady
	// StatusRunning is when the store has been successfully bootstrapped.
	StatusRunning
)

func (s Status) String() string {
	switch s {
	case StatusInit:
		return "init"
	case StatusSetup:
		return "setup"
	case StatusReady:
		return "ready"
	case StatusRunning:
		return "running"
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

	// Ensure that required directories exist. This also involves creating a blank
	// directory for the root directory just for the sake of completion.
	dirs := []string{"", "raft"}
	for _, dir := range dirs {
		path := filepath.Join(e.DataPath, dir)
		_, err := os.Stat(path)
		if !os.IsNotExist(err) {
			continue
		}
		log.Printf("[INFO] engine: Creating new directory %s", path)
		os.MkdirAll(path, 0644)
	}

	// Read in the config
	if err := e.readConfig(); err != nil {
		return err
	}

	// Start the API server.
	go func() { errCh <- e.APIServer.Start() }()

	// If the engine is ready, start the RPC server and the store.
	if e.Status >= StatusReady {
		go func() { errCh <- e.RPCServer.Start() }()
		go func() { errCh <- e.Store.Open() }()
	}

	// Monitor started progress on each component.
	go func() {
		<-e.APIServer.Started()

		if e.Status >= StatusReady {
			<-e.Store.Started()
			<-e.RPCServer.Started()
		}

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

// Reset will reset all engine properties and wipe all local data files.
func (e *Engine) Reset() error {
	return nil
}

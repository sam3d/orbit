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
	// StatusReady is when the engine has been successfully bootstrapped, but
	// before it has been fully configured with a domain name or user.
	StatusReady
	// StatusRunning is when the store has been successfully bootstrapped and a
	// user has set themselves up fully.
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

// SetupStatus is a string representation of the stage and mode.
func (e *Engine) SetupStatus() (mode, stage string) {
	// If the engine is in running mode, then there is nothing to do.
	if e.Status == StatusRunning {
		return "complete", "complete"
	}

	// If the engine is in setup mode, then there is nothing to do.
	if e.Status == StatusSetup {
		return "bootstrap", "welcome"
	}

	// If the engine is not ready, don't do anything. It's only if the engine is
	// ready that all of the following conditions apply about the setup location.
	if e.Status != StatusReady {
		return
	}

	// If there is only one node that the cluster is aware of, it means that this
	// node must be the one responsible for establishing the cluster. Otherwise,
	// it means that this node must be joining the cluster, which means that
	// because the engine is not running, they must be in the node configuration stage.
	if len(e.Store.state.Nodes) > 1 {
		return "join", "node"
	}

	// If there are no routers, that means that we must need to set up the domain
	// that is used for routing all Orbit traffic.
	if len(e.Store.state.Routers) == 0 {
		return "bootstrap", "domain"
	}

	if len(e.Store.state.Users) == 0 {
		return "bootstrap", "user"
	}

	// If the single node that is in the cluster does not have any roles, we can
	// assume that this hasn't yet been configured and so this is the final stage
	// of the system.
	if !e.Store.state.Nodes[0].HasRole(RoleManager) {
		return "bootstrap", "node"
	}

	// Otherwise, the store state must be complete.
	return "bootstrap", "complete"
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

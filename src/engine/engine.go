// Package engine provides the all-encompassing interface to the Orbit
// background operations. This includes replicated state management, gossip
// control, and ensuring that the state is maintained for the respective nodes.
package engine

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net"
	"os"
	"path/filepath"
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
)

func (s Status) String() string {
	switch s {
	case Init:
		return "init"
	case Setup:
		return "setup"
	default:
		return ""
	}
}

// Start starts the engine and all of its subcomponents. This is dependent on
// state, so for example if the cluster still has yet to be set up, then it
// won't start the store.
func (e *Engine) Start() error {
	log.Println("[INFO] engine: Starting...")

	err := make(chan error) // Main error channel closure

	// Read in the config
	if err := e.readConfig(); err != nil {
		return err
	}

	// Start the API Server.
	go func() {
		err <- e.APIServer.Start()
	}()

	// Monitor started progress on each component.
	go func() {
		<-e.APIServer.Started()
		log.Println("[INFO] engine: Started")
	}()

	return <-err
}

// Stop will stop the operation of the engine instance.
func (e *Engine) Stop() error {
	log.Println("[INFO] engine: Stopping...")
	log.Println("[INFO] engine: Stopped")
	return nil
}

// config is the configuration struct that the engine saves and interacts with.
type config struct {
	AdvertiseAddr string `json:"advertise_addr"`
	Status        Status `json:"status"`
}

// createConfig will create the configuration file for the engine.
func (e *Engine) createConfig() error {
	path := filepath.Join(e.DataPath, e.ConfigFile)

	defaultConfig := &config{Status: Setup} // By default, we want to setup mode
	b, err := json.MarshalIndent(defaultConfig, "", "  ")
	if err != nil {
		return err
	}
	b = append(b, '\n') // Add newline to the end of the file

	if err := ioutil.WriteFile(path, b, 0644); err != nil {
		return err
	}

	return nil
}

// readConfig will read in the configuration file and parse it.
func (e *Engine) readConfig() error {
	path := filepath.Join(e.DataPath, e.ConfigFile)

	file, err := ioutil.ReadFile(path)
	if err != nil {
		// Check if the error is that the file does not exist so that we can create
		// it.
		if os.IsNotExist(err) {
			log.Printf("[INFO] engine: Creating %s\n", path)
			if err := e.createConfig(); err != nil {
				log.Printf("[ERR] engine: Could not create %s\n", path)
				return err
			}
			// Now that we have re-created the config file, we can create it.
			return e.readConfig()
		}

		// Check if the file can't be read.
		if err == os.ErrPermission {
			log.Printf("[ERR] engine: Insufficient read permissions for %s\n", path)
			return err
		}

		// It's none of the above cases, so just return the error.
		return err
	}

	// Put the config file in the config struct
	var config config
	if err := json.Unmarshal(file, &config); err != nil {
		log.Printf("[ERR] engine: Parsing config file %s\n", path)
		return err
	}

	// And now we actually set the config file contents
	e.Store.AdvertiseAddr = net.ParseIP(config.AdvertiseAddr)
	e.Status = config.Status

	log.Printf("[INFO] engine: Imported config %s\n", path)
	return nil
}

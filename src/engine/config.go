package engine

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"path/filepath"
)

// config is the configuration struct that the engine saves and interacts with.
type config struct {
	AdvertiseAddr string `json:"advertise_addr"`
	Status        Status `json:"status"`
	ID            string `json:"id"`
}

// configPath will return the path of the config file that the engine will use.
// By default, this is a concatenation of the DataPath and ConfigFile struct
// fields.
func (e Engine) configPath() string {
	return filepath.Join(e.DataPath, e.ConfigFile)
}

// createConfig will create the configuration file for the engine.
func (e Engine) createConfig() error {
	path := e.configPath()

	defaultConfig := &config{Status: Setup} // By default, we want to setup mode
	b, err := json.MarshalIndent(defaultConfig, "", "  ")
	if err != nil {
		return err
	}
	b = append(b, '\n') // Add newline to the end of the file

	if err := ioutil.WriteFile(path, b, 0666); err != nil {
		return err
	}

	return nil
}

// readConfig will read in the configuration file and parse it.
func (e *Engine) readConfig() error {
	path := e.configPath()

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

	// The config status should never be 0 in the config file (init), as this is a
	// status reserved for before the cluster has loaded the config. Now that is
	// *has* loaded, set it to 1 (setup), or the first state for a cluster after
	// it has been setup.
	if config.Status == 0 {
		config.Status = 1
	}

	// And now we actually set the config file contents
	e.Store.AdvertiseAddr = net.ParseIP(config.AdvertiseAddr)
	e.Status = config.Status
	e.Store.ID = config.ID

	log.Printf("[INFO] engine: Imported config %s\n", path)

	// Perform test write that doesn't change any of the data. This will format
	// the data in the file if that hasn't been correctly formatted up until
	// this point.
	if err := e.writeConfig(); err != nil {
		return err
	}

	return nil
}

// writeConfig will create a config file based on the current engine settings.
func (e Engine) writeConfig() error {
	path := e.configPath()

	file, err := os.Create(path)
	if err != nil {
		log.Printf("[ERR] engine: Could not open config for writing %s\n", path)
		return err
	}
	defer file.Close()

	// Ensure that the address is not null. If it is, make sure to use a blank
	// string.
	var parsedAddr string
	if e.Store.AdvertiseAddr == nil {
		parsedAddr = ""
	} else {
		parsedAddr = fmt.Sprintf("%s", e.Store.AdvertiseAddr)
	}

	config := config{
		Status:        e.Status,
		AdvertiseAddr: parsedAddr,
		ID:            e.Store.ID,
	}

	en := json.NewEncoder(file)
	en.SetIndent("", "  ")
	if err := en.Encode(&config); err != nil {
		log.Printf("[ERR] engine: Could not write config: %s\n", path)
		return err
	}

	log.Printf("[INFO] engine: Updated config %s", path)
	return nil
}

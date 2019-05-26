package engine

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
)

// Deployment is a store instance of a deployment created from an image or
// repository. It could also be referred to as an "App" (and in some cases
// throughout Orbit, actually is).
type Deployment struct {
	ID   string `json:"id"`
	Name string `json:"name"`

	// The location that the deployment is created from.
	Repository string `json:"repository"`
	Path       string `json:"path"` // A subdirectory or root of the repo

	// The logs from the build processes. This is a map that contains a string
	// (the key) which is used to store the git commit hash of the repository that
	// this particular deployment path was taken from. The value is a string list
	// of the individual lines outputted from the build process. These all need to
	// be kept in raft consensus so that they can be referenced later on.
	BuildLogs map[string][]string

	NamespaceID string `json:"namespace_id"`
}

// Deployments is a slice of the deployments in the store.
type Deployments []Deployment

// BuildDeployment will take in the given deployment object and then run through
// and actually perform the operations to build that deployment.
func (e *Engine) BuildDeployment(d Deployment) error {
	// Checkout the repo to a temporary directory, navigate to the specified path,
	// and if there is a Dockerfile, use that for building, and if there isn't,
	// create a default one that uses the herokuish image.

	// Find the repo.
	var repo *Repository
	for _, r := range e.Store.state.Repositories {
		if r.ID == d.Repository {
			repo = &r
			break
		}
	}
	if repo == nil {
		return fmt.Errorf("that repository does not exist")
	}

	// Derive the repo path.
	volume := e.Store.OrbitSystemVolume()
	if volume == nil {
		return fmt.Errorf("could not find the orbit system volume")
	}
	path := filepath.Join(volume.Paths().Data, "repositories", repo.ID)

	// Check it out to a temporary directory.
	tmp, err := ioutil.TempDir("", "")
	if err != nil {
		return fmt.Errorf("could not create temporary directory: %s", err)
	}
	cmd := exec.Command("git", "clone", path, tmp)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("could not run git clone command: %s", err)
	}

	return nil
}

// GenerateID will create a unique identifier for the deployment.
func (d *Deployments) GenerateID() string {
search:
	for {
		b := make([]byte, 8)
		rand.Read(b)
		id := hex.EncodeToString(b)

		for _, deployment := range *d {
			if deployment.ID == id {
				continue search
			}
		}

		return id
	}
}

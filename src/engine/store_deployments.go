package engine

import (
	"crypto/rand"
	"encoding/hex"
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

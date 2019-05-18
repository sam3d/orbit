package engine

import (
	"crypto/rand"
	"encoding/hex"
)

// Repository is the information about the code that someone has on the system.
// This is generally a git repository, and it is used by the API git component
// to decide where to put the code.
type Repository struct {
	ID   string `json:"id"`
	Name string `json:"name"`

	NamespaceID string `json:"namespace_id"`
}

// Repositories is a group of the store repository. The group allows for easier
// ID generation.
type Repositories []Repository

// GenerateID will create a unique identifier for the repository.
func (r *Repositories) GenerateID() string {
search:
	for {
		b := make([]byte, 8)
		rand.Read(b)
		id := hex.EncodeToString(b)

		for _, repo := range *r {
			if repo.ID == id {
				continue search
			}
		}

		return id
	}
}

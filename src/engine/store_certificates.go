package engine

import (
	"encoding/hex"
	"math/rand"
)

// Certificate is a TLS certificate.
type Certificate struct {
	ID        string `json:"id"`
	Data      []byte `json:"data"`       // The raw certificate data
	AutoRenew bool   `json:"auto_renew"` // Whether or not to auto renew cert
}

// Certificates is a group of TLS certificates.
type Certificates []Certificate

// GenerateID will create a new certificate ID based on the existing
// certificates.
func (c *Certificates) GenerateID() string {
search:
	for {
		b := make([]byte, 32)
		rand.Read(b)
		id := hex.EncodeToString(b)

		for _, user := range *c {
			if user.ID == id {
				continue search
			}
		}

		return id
	}
}

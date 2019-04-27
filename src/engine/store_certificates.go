package engine

import (
	"encoding/hex"
	"math/rand"
)

// Certificate is a TLS certificate.
type Certificate struct {
	ID          string `json:"id"`
	AutoRenew   bool   `json:"auto_renew"`   // Whether or not to auto renew cert
	NamespaceID string `json:"namespace_id"` // The ID of the namespace

	FullChain  []byte `json:"full_chain"`  // The full chain certificate data
	PrivateKey []byte `json:"private_key"` // The private key of the certificate
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

		for _, certificate := range *c {
			if certificate.ID == id {
				continue search
			}
		}

		return id
	}
}

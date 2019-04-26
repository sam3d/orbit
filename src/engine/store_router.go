package engine

import (
	"encoding/hex"
	"math/rand"
)

// Router is a routing point on the store. This is used for domain names, ports,
// and paths. It has a many-to-one relationship with certificates: a domain name
// can only use one certificate, but a certificate can be used with many
// different domain names (especially if it's a wildcard).
type Router struct {
	ID            string `json:"id"`
	Domain        string `json:"domain"`
	CertificateID string `json:"certificate_id"`
	NamespaceID   string `json:"namespace_id"`
}

// Routers is a group of domain names, ports, and paths, used for routing.
type Routers []Router

// GenerateID will create a new router ID based on the existing ID's.
func (r *Routers) GenerateID() string {
search:
	for {
		b := make([]byte, 32)
		rand.Read(b)
		id := hex.EncodeToString(b)

		for _, router := range *r {
			if router.ID == id {
				continue search
			}
		}

		return id
	}
}

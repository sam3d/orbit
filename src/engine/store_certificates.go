package engine

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"

	"golang.org/x/crypto/acme"
)

// Challenge is a HTTP-01 letsencrypt challenge. It contains the path, token,
// and domain required for successfully serving the challenge response.
type Challenge struct {
	Path   string `json:"path"`
	Token  string `json:"token"`
	Domain string `json:"domain"`
}

// Certificate is a TLS certificate.
type Certificate struct {
	ID          string      `json:"id"`
	AutoRenew   bool        `json:"auto_renew"`   // Whether or not to auto renew cert
	Domains     []string    `json:"domains"`      // The domain names for which the certificate is valid
	Challenges  []Challenge `json:"challenges"`   // Pending challenges for this certificate
	NamespaceID string      `json:"namespace_id"` // The ID of the namespace

	FullChain  []byte `json:"full_chain"`  // The full chain certificate data
	PrivateKey []byte `json:"private_key"` // The private key of the certificate
}

// Certificates is a group of TLS certificates.
type Certificates []Certificate

func newACMEClient() (*acme.Client, error) {
	acctKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		log.Printf("[ERR] renew: Could not generate private key: %s", err)
		return nil, err
	}

	client := &acme.Client{
		Key:          acctKey,
		DirectoryURL: "https://acme-staging.api.letsencrypt.org/directory",
	}

	_, err = client.Register(context.Background(), &acme.Account{}, func(tos string) bool { return true })
	if err != nil {
		log.Printf("[ERR] renew: Could not create account: %s", err)
		return nil, err
	}

	return client, nil
}

func Renew(cert Certificate) {
	// Create ACME client.
	client, err := newACMEClient()
	if err != nil {
		return
	}

	// Keep track of the authorizations.

	for _, domain := range cert.Domains {
		// Create the authorization for this domain.
		auth, err := client.Authorize(context.Background(), domain)
		if err != nil {
			log.Printf("[ERR] renew: Could not authorize: %s", err)
			return
		}

		// Ensure that the challenge exists and is valid.
		var challenge *acme.Challenge
		for _, c := range auth.Challenges {
			if c.Type == "http-01" {
				challenge = c
				break
			}
		}
		if challenge == nil {
			log.Print("[ERR] renew: No HTTP-01 challenge present")
			return
		}

		// Retrieve the challenge properties.
		path := client.HTTP01ChallengePath(challenge.Token)
		res, err := client.HTTP01ChallengeResponse(challenge.Token)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(path, "\n\n", res, "\n\n\n")
	}
}

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

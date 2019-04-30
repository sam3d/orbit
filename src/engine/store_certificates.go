package engine

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/hex"
	"log"

	"github.com/pkg/errors"
	"golang.org/x/crypto/acme"
)

// Certificate is a TLS certificate.
type Certificate struct {
	ID          string   `json:"id"`
	Domains     []string `json:"domains"`      // The domain names for which the certificate is valid
	NamespaceID string   `json:"namespace_id"` // The ID of the namespace

	FullChain  []byte `json:"full_chain"`  // The full chain certificate data
	PrivateKey []byte `json:"private_key"` // The private key of the certificate

	AutoRenew  bool        `json:"auto_renew"` // Whether or not to auto renew cert
	Challenges []Challenge `json:"challenges"` // Pending challenges for this certificate
}

// Certificates is a group of TLS certificates.
type Certificates []Certificate

// Challenge is a HTTP-01 letsencrypt challenge. It contains the path, token,
// and domain required for successfully serving the challenge response.
type Challenge struct {
	Path   string `json:"path"`
	Token  string `json:"token"`
	Domain string `json:"domain"`
}

// RenewCertificates will undergo the issuance and distributed of the
// certificate challenges, and then update the certificates should that be
// required.
func (s *Store) RenewCertificates() error {
	// Create the ACME client.
	client, err := newACMEClient()
	if err != nil {
		return errors.Wrap(err, "could not create ACME client")
	}

	// Retrieve the challenges for the certificates.
	for _, cert := range s.state.Certificates {
		// Skip the auto-renewal for this certificate if it is not enabled. This
		// means that no LetsEncrypt operations will take place unless this has been
		// specified.
		if !cert.AutoRenew {
			continue
		}

		// Keep track of a list of challenges for the domains in this certificate.
		var challenges []Challenge

		// Loop over all of the domains in this certificate and prepare challenges
		// for each of them.
		for _, domain := range cert.Domains {
			auth, err := client.Authorize(context.Background(), domain)
			if err != nil {
				return errors.Wrap(err, "could not authorize the domain")
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
				return errors.Wrap(err, "no http-01 challenge present")
			}

			// Retrieve the challenge properties.
			path := client.HTTP01ChallengePath(challenge.Token)
			res, err := client.HTTP01ChallengeResponse(challenge.Token)
			if err != nil {
				return errors.Wrap(err, "could not retrieve challenge token")
			}

			// Construct the challenge type and append it to the list of challenges.
			challenges = append(challenges, Challenge{
				Path:   path,
				Token:  res,
				Domain: domain,
			})
		}

		// Now we have the challenges for this certificate, let's create and apply
		// the challenges to the certificate object in the store.
		cmd := command{
			Op: opUpdateCertificate,
			Certificate: Certificate{
				ID:         cert.ID,
				Challenges: challenges,
			},
		}

		if err := cmd.Apply(s); err != nil {
			return errors.Wrap(err, "could not add challenges to certificate")
		}
	}

	// All of the certificates now have challenges on them, update the load
	// balancers to perform the update.
	// TODO: Implement this.

	return nil
}

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

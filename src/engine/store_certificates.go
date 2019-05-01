package engine

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"log"

	"github.com/pkg/errors"
	"golang.org/x/crypto/acme"
	"orbit.sh/engine/docker"
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

	// Request is the overarching structure to keep track of the certificates
	// during the renewal process.
	type Request struct {
		Certificate    *Certificate
		Challenges     []*acme.Challenge
		Authorizations []*acme.Authorization
		Errors         []error
	}

	// requests is for keeping track of the certificate requests.
	var requests []*Request

	// Retrieve the challenges for the certificates.
	for _, cert := range s.state.Certificates {
		// Skip the auto-renewal for this certificate if it is not enabled. This
		// means that no LetsEncrypt operations will take place unless this has been
		// specified.
		if !cert.AutoRenew {
			continue
		}

		// Prepare the overall request object for this certificate.
		req := &Request{Certificate: &cert}
		requests = append(requests, req)

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

			// Keep track of the acme challenge so that they can be accepted.
			req.Challenges = append(req.Challenges, challenge)

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
	// balancers to start serving the LetsEncrypt challenges.
	if err := docker.ForceUpdateService("edge"); err != nil {
		return errors.Wrap(err, "could not restart the edge routers")
	}

	// Now that the load balancers have been updated, we can perform all of the
	// operations required to actually requisition the certificates.

	// Accept all of the challenges.
	for _, r := range requests {
		for _, c := range r.Challenges {
			if _, err := client.Accept(context.Background(), c); err != nil {
				return errors.Wrap(err, "could not confirm acceptance of challenge")
			}
		}
	}

	// Authorize all of the challenges.
	for _, r := range requests {
		for _, c := range r.Challenges {
			auth, err := client.WaitAuthorization(context.Background(), c.URI)
			if err != nil {
				// If an authorization fails, we can't make the certificate request. All
				// of the authorizations need to succeed in a certificate to be able to
				// undertake certificate retrieval.
				log.Printf("[ERR] certs: Could not authorize certificate (%s)", c.URI)
				r.Errors = append(r.Errors, err)
				continue
			}

			// The authorization succeeded, keep track of it.
			r.Authorizations = append(r.Authorizations, auth)
		}
	}

	// For each of requests with successful certificates, construct the
	// certificate signing requests and issue them.
	for _, r := range requests {
		if len(r.Errors) > 0 {
			log.Printf("[ERR] certs: Skipping certificate %s, as %d of the domains encountered errors", r.Certificate.ID, len(r.Errors))
			continue
		}

		// Generate a private key for this certificate.
		certKey, err := rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			return errors.Wrap(err, "could not generate private key")
		}

		// Construct the certificate request.
		req := &x509.CertificateRequest{
			Subject:  pkix.Name{CommonName: r.Certificate.Domains[0]},
			DNSNames: r.Certificate.Domains,
		}

		// Create the actual signing request.
		csr, err := x509.CreateCertificateRequest(rand.Reader, req, certKey)
		if err != nil {
			return errors.Wrap(err, "could not create certificate request")
		}

		// Create the encoded certificate with the signing request.
		der, url, err := client.CreateCert(context.Background(), csr, 0, true)
		if err != nil {
			return errors.Wrap(err, "could not create certificate")
		}
		fmt.Println("URL:", url)

		// A full chain certificate is simply multiple certificates appended
		// together. Important to note is that you don't just append the bytes, but
		// you literally have a single file with multiple "--BEGIN CERTIFICATE--"
		// and "--END CERTIFICATE--" blocks just one after the other. That's what
		// this next bit does.
		var cert []byte
		for _, b := range der {
			block := pem.EncodeToMemory(&pem.Block{
				Type:  "CERTIFICATE",
				Bytes: b,
			})
			cert = append(cert, block...)
		}

		// Now we need to convert the private key to PEM format.
		derCertKey := x509.MarshalPKCS1PrivateKey(certKey)
		pemCertKey := pem.EncodeToMemory(&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: derCertKey,
		})

		// Update the certificate with the new LetsEncrypt certificate data.
		cmd := command{
			Op: opUpdateCertificate,
			Certificate: Certificate{
				ID:         r.Certificate.ID,
				FullChain:  cert,
				PrivateKey: pemCertKey,
			},
		}
		if err := cmd.Apply(s); err != nil {
			return errors.Wrap(err, "could not apply certificate update to store")
		}
	}

	// All of the certificates have been updated, let's do one final reload of the
	// load balancers to intake the updated certificates.
	if err := docker.ForceUpdateService("edge"); err != nil {
		return errors.Wrap(err, "could not update edge routers")
	}

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

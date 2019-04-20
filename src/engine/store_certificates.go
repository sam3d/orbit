package engine

// Certificate is a TLS certificate.
type Certificate struct {
	ID string `json:"id"`
}

// Certificates is a group of TLS certificates.
type Certificates []Certificate

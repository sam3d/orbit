package engine

// Router is a routing point on the store. This is used for domain names, ports,
// and paths. It has a many-to-one relationship with certificates: a domain name
// can only use one certificate, but a certificate can be used with many
// different domain names (especially if it's a wildcard).
type Router struct {
	ID          string       `json:"id"`
	Domain      string       `json:"domain"`
	Certificate *Certificate `json:"certificate"`
}

// Routers is a group of domain names, ports, and paths, used for routing.
type Routers []Router

package main

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"strings"
)

// Client is a way of interacting with the Orbit unix socket.
type Client struct {
	client *http.Client
}

// NewClient creates a new instance of the orbit socket client.
func NewClient() *Client {
	return &Client{
		client: &http.Client{
			Transport: &http.Transport{
				DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
					return net.Dial("unix", "/var/run/orbit.sock")
				},
			},
		},
	}
}

// Get makes a GET request to the Orbit socket.
func (c *Client) Get(path string) []byte {
	url := "http://unix/" + strings.TrimPrefix(path, "/")
	res, err := c.client.Get(url)
	if err != nil {
		log.Fatalf("Could not query Orbit socket: %s", err)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatalf("Could not read HTTP response from Orbit socket: %s", err)
	}

	return body
}

// Router is a logical Orbit router.
type Router struct {
	ID            string `json:"id"`
	Domain        string `json:"domain"`
	AppID         string `json:"app_id"`
	CertificateID string `json:"certificate_id"`
	WWWRedirect   bool   `json:"www_redirect"`
}

// Certificate is a logical certificate
type Certificate struct {
	ID         string `json:"id"`
	FullChain  []byte `json:"full_chain"`
	PrivateKey []byte `json:"private_key"`
}

// GetRouters retrieves all of the routers from the Orbit socket.
func (c *Client) GetRouters() []Router {
	body := c.Get("/routers")
	routers := []Router{}

	if err := json.Unmarshal(body, &routers); err != nil {
		log.Fatalf("Could not parse response from Orbit socket: %s", err)
	}

	return routers
}

// GetCertificates retrieves all of the certificates from the Orbit socket.
func (c *Client) GetCertificates() []Certificate {
	body := c.Get("/certificates")
	certificates := []Certificate{}

	if err := json.Unmarshal(body, &certificates); err != nil {
		log.Fatalf("Could not parse response from Orbit socket: %s", err)
	}

	return certificates
}

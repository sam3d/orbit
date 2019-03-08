package engine

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
)

// RPCClient is a client for connecting to, making, and using RPC requests.
type RPCClient struct {
	Addr net.TCPAddr
}

// RPCResponse is a response object from an RPC client.
type RPCResponse struct {
	StatusCode int
	Status     string
	Body       []byte
}

// Post will make a request to the previously allocated RPC target.
func (c *RPCClient) Post(path string, data interface{}) (*RPCResponse, error) {
	b, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	target := "http://" + c.Addr.String() + "/" + strings.TrimPrefix(path, "/")
	resp, err := http.Post(target, "application/json", bytes.NewBuffer(b))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return &RPCResponse{
		StatusCode: resp.StatusCode,
		Status:     resp.Status,
		Body:       body,
	}, nil
}

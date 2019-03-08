package engine

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net"
	"net/http"
	"strings"

	"github.com/pkg/errors"
)

// RPCClient is a client for connecting to, making, and using RPC requests.
type RPCClient struct {
	Addr net.TCPAddr
}

// Post will make a request to the previously allocated RPC target.
func (c *RPCClient) Post(path string, data interface{}, obj interface{}) (*http.Response, error) {
	b, err := json.Marshal(data)
	if err != nil {
		return nil, errors.Wrap(err, "could not marshal data into json")
	}

	target := "http://" + c.Addr.String() + "/" + strings.TrimPrefix(path, "/")
	resp, err := http.Post(target, "application/json", bytes.NewBuffer(b))
	if err != nil {
		return nil, errors.Wrapf(err, "request to %s failed", target)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return resp, errors.Wrap(err, "could not ready from response body")
	}
	resp.Body = ioutil.NopCloser(bytes.NewReader(body)) // Reset buffer.
	if err := json.Unmarshal(body, obj); err != nil {
		return resp, errors.Wrap(err, "could not marshal json into object interface")
	}

	return resp, nil
}

// RPCJoinRequest is sent from a joining node to a target node when it wishes to
// participate in the cluster.
type RPCJoinRequest struct {
	JoinToken string `json:"join_token"`
}

// RPCJoinResponse is sent back from a target node to the joining node.
type RPCJoinResponse struct {
	// The following properties are for the joining node to use.
	AdvertiseAddr string `json:"advertise_address"`
	ID            string `json:"id"`
}

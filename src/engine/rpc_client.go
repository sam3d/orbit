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
	// Convert anonymous data object into JSON.
	b, err := json.Marshal(data)
	if err != nil {
		return nil, errors.Wrap(err, "could not marshal data into json")
	}

	// Actually make the request.
	target := "http://" + c.Addr.String() + "/" + strings.TrimPrefix(path, "/")
	resp, err := http.Post(target, "application/json", bytes.NewBuffer(b))
	if err != nil {
		return nil, errors.Wrapf(err, "request to %s failed", target)
	}

	// Read from the request, and then reset the request object if needed.
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return resp, errors.Wrap(err, "could not read from response body")
	}
	resp.Body = ioutil.NopCloser(bytes.NewReader(body))

	// Place the response data into the anonymous object.
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

	// The following is the information from the remote node.
	RaftPort    int `json:"raft_port"`
	SerfPort    int `json:"serf_port"`
	WANSerfPort int `json:"wan_serf_port"`
}

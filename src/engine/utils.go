package engine

import (
	"io/ioutil"
	"net"
	"net/http"

	"github.com/pkg/errors"
)

// getPublicIP will retrieve the public IP for this node.
func getPublicIP() (net.IP, error) {
	res, err := http.Get("https://api.ipify.org")
	if err != nil {
		return nil, errors.Wrap(err, "could not connect to ipify API")
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, errors.Wrap(err, "could not read ipify response body")
	}

	ip := net.ParseIP(string(body))
	if ip == nil {
		return nil, errors.New("could not parse IP address")
	}

	return ip, nil
}

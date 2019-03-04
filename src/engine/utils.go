package engine

import (
	"io/ioutil"
	"net/http"

	"github.com/pkg/errors"
)

// getPublicIP will retrieve the public IP for this node.
func getPublicIP() (string, error) {
	res, err := http.Get("https://api.ipify.org")
	if err != nil {
		return "", errors.Wrap(err, "could not connect to ipify API")
	}

	ip, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", errors.Wrap(err, "could not read ipify response body")
	}

	return string(ip), nil
}

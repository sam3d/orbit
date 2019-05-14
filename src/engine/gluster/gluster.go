// Package gluster implements basic tools for interacting with the GlusterFS
// command line interface.
package gluster

import (
	"log"
	"os"
	"os/exec"
)

// PeerProbe performs a peer probe on an address. This only works if the node is
// not already in a cluster, and will fail if it is.
func PeerProbe(ip string) error {
	cmd := exec.Command("gluster", "peer", "probe", ip)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Printf("[ERR] gluster: Could not perform peer probe on %s: %s", ip, err)
		return err
	}
	return nil
}

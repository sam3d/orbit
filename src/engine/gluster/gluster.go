// Package gluster implements basic tools for interacting with the GlusterFS
// command line interface.
package gluster

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
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

// Fallocate will create a block at the given path with the given size in
// megabytes.
func Fallocate(path string, size int) error {
	length := fmt.Sprintf("%dMB", size)
	cmd := exec.Command("fallocate", "--length", length, path)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

// MakeFS will take in the path of a block volume and make it into an filesystem
// of the given type.
func MakeFS(filesystem, path string) error {
	bin := fmt.Sprintf("mkfs.%s", strings.ToLower(filesystem))
	cmd := exec.Command(bin, path)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

// Mount will run the mount command on a simple path.
func Mount(from, to string) error {
	cmd := exec.Command("mount", from, to)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

// MountGluster will mount a GlusterFS volume from a specified path.
func MountGluster(ip, volume, to string) error {
	from := fmt.Sprintf("%s:/%s", ip, volume)
	cmd := exec.Command("mount", "-t", "glusterfs", from, to)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

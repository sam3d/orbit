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

// ExistingMount is a struct for when a mount is already present.
type ExistingMount struct {
	From string `json:"from"`
	To   string `json:"to"`
}

// ExistingMounts returns the mounts that are already present on the server.
func ExistingMounts() ([]ExistingMount, error) {
	// Retrieve the command output for the "mount" command.
	cmd := exec.Command("mount")
	cmd.Stderr = os.Stderr
	output, err := cmd.Output()
	if err != nil {
		// There was an error running the command that retrieves the mounts. Return
		// an empty slice.
		return nil, err
	}

	var mounts []ExistingMount

	// Tokenise the response so that we can figure out which mounts we need to
	// use.
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if !strings.Contains(line, " on ") {
			// Isn't a mount path we can check, so ignore it completely.
			break
		}

		// Retrieve the correct components of the mount path and append it to the
		// existing mount slice by creating the mount object.
		tokens := strings.Split(line, " ")
		mounts = append(mounts, ExistingMount{
			From: tokens[0],
			To:   tokens[2],
		})
	}

	return mounts, nil
}

// AlreadyMounted returns if a volume has already been mounted.
func AlreadyMounted(from, to string) bool {
	// Retrieve the existing mounted directories.
	mounts, err := ExistingMounts()
	if err != nil {
		log.Printf("[ERR] gluster: Can't retrieve the existing mounts: %s", err)
		return false
	}

	// Check the provided mount against the existing mounts.
	for _, m := range mounts {
		if m.From == from && m.To == to {
			// The mount does exist, so we can continue.
			return true
		}
	}

	// The mount must not exist.
	return false
}

// Mount will run the mount command on a simple path.
func Mount(from, to string) error {
	// Ensure that it's not already mounted.
	if AlreadyMounted(from, to) {
		return nil
	}

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

	// Check whether this volume is already mounted, and if so, do nothing.
	if AlreadyMounted(from, to) {
		return nil
	}

	cmd := exec.Command("mount", "-t", "glusterfs", from, to)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

// Mode the replication mode.
type Mode uint

const (
	// Replica is a replicated data mode.
	Replica Mode = iota
)

// CreateVolume will create a gluster volume.
func CreateVolume(id string, bricks []string, mode Mode) error {
	args := []string{"volume", "create", id} // Create the initial args

	// Set the replica mode as long as there is more than one and the mode is nothing or replica.
	if (mode == Replica) && len(bricks) > 1 {
		count := len(bricks)
		args = append(args, "replica", string(count))
	}

	// Construct the command.
	args = append(args, bricks...) // Append all of the brick strings to the args
	cmd := exec.Command("gluster", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Log the command out so we can debug.
	log.Printf("[INFO] gluster: Running command '%s'", strings.Join(args, " "))

	// Run the command.
	if err := cmd.Run(); err != nil {
		log.Printf("[ERR] gluster: Could not run volume create command on volume %s: %s", id, err)
		return err
	}

	return nil
}

// StartVolume will start a glusterfs volume by ID.
func StartVolume(id string) error {
	cmd := exec.Command("gluster", "volume", "start", id)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Printf("[ERR] gluster: Could not run volume start command on volume %s: %s", id, err)
		return err
	}
	return nil
}

package engine

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"path/filepath"
	"time"

	"orbit.sh/engine/gluster"
)

// Volume is a distributed block storage volume propagated by GlusterFS.
type Volume struct {
	ID     string  `json:"id"`
	Name   string  `json:"name"`   // The short (friendly) name for the volume
	Size   int     `json:"size"`   // Size of volume in MB (used for allocation)
	Bricks []Brick `json:"bricks"` // The different bricks for this volume

	NamespaceID string `json:"namespace_id"`
}

// Brick is a server and a location for a given bit of data. It keeps track of
// whether or not it's been created so that the target node can inform the
// leader of the cluster as to whether not it's safe to progress to gluster
// volume creation (the "volume create" command will fail on the respective
// nodes if this is the case).
type Brick struct {
	NodeID  string `json:"node_id"` // The ID of the node hosting the block
	Created bool   `json:"created"` // Set by target node, whether or not it's been created
}

// Volumes is a list of volumes in the store.
type Volumes []Volume

// RootVolumeDir is the root filesystem where the volumes are stored. This
// will be the same for each volume.
const RootVolumeDir = "/var/orbit/volumes"

// VolumePaths returns the data about where the absolute (read: not relative)
// paths for a given volume should be. This makes the handling of volume paths
// more consistent.
type VolumePaths struct {
	// Container is the overall location for that volume data.
	Container string `json:"container"`
	// Data is where the started gluster volume should be mounted.
	Data string `json:"data"`
	// Raw is the raw block storage filesystem to be allocated.
	Raw string `json:"raw"`
	// Volume is where the raw block gets mounted. It isn't used by gluster
	// directly (it's not recommended to use the root of a filesystem for a
	// gluster mount). Instead the "brick" subdirectory gets used by gluster.
	Volume string `json:"volume"`
	// Brick is the directory inside of volume directory that gluster uses as its
	// actual brick.
	Brick string `json:"brick"`
}

// Paths returns the desired paths for a given volume.
func (v Volume) Paths() VolumePaths {
	container := filepath.Join(RootVolumeDir, v.ID)

	raw := filepath.Join(container, "raw")
	volume := filepath.Join(container, "volume")
	data := filepath.Join(container, "data")
	brick := filepath.Join(volume, "brick")

	return VolumePaths{
		Container: container,
		Raw:       raw,
		Volume:    volume,
		Brick:     brick,
		Data:      data,
	}
}

// AddVolume is a shorthand method on the store that will go through all of the
// steps required for making a volume. The ID is auto generated if none is
// provided when this is used, so only the non-auto-generated properties need to
// be used.
func (s *Store) AddVolume(v Volume) (*Volume, error) {
	// Auto generate a volume ID if none was provided.
	if v.ID == "" {
		v.ID = s.state.Volumes.GenerateID()
	}

	// Create the new volume command.
	cmd := command{
		Op:     opNewVolume,
		Volume: v,
	}

	// Apply it to the store.
	if err := cmd.Apply(s); err != nil {
		log.Printf("[ERR] volume: Could not apply the volume to the store: %s", err)
		return &v, err
	}

	// Wait for all of the nodes to create the volume data for themselves after
	// propagation.
	s.WaitForVolume(v.ID)

	// Derive the paths to use for the bricks.
	paths := cmd.Volume.Paths()
	var bricks []string
	for _, b := range cmd.Volume.Bricks {
		// Find the node for that brick.
		for _, n := range s.state.Nodes {
			if b.NodeID == n.ID {
				// Add the path and continue.
				bricks = append(bricks, fmt.Sprintf("%s:%s", n.Address, paths.Brick))
				break
			}
		}
	}

	// Now create and start the volume.
	gluster.CreateVolume(v.ID, bricks, gluster.Replica)
	gluster.StartVolume(v.ID)

	return &v, nil
}

// WaitForVolume will wait for the volume and all of its respective bricks to be
// created. This does not perform the "gluster volume create" operation, as that
// only needs to be run on a single node.
func (s *Store) WaitForVolume(id string) {
search:
	for {
		// No matter what, at the beginning of this search, we always wait 0.2
		// seconds. This is to prevent blocking too much and also to ensure that we
		// wait before checking after either an incomplete volume find or an
		// incomplete brick creation.
		time.Sleep(time.Millisecond * 200)

		// Find the volume. This is repeated as the reference can continuously
		// update and we need to ensure we have access to the latest bits of data.
		var volume *Volume
		for _, v := range s.state.Volumes {
			if v.ID == id {
				volume = &v
				break
			}
		}

		// If there is no volume that matches this description, it means it does not
		// yet exist on the store. We need to continue the search.
		if volume == nil {
			continue search
		}

		// Check the bricks in the volume.
		for _, b := range volume.Bricks {
			if !b.Created {
				// At least one brick in this volume has not been created. Start the
				// whole search again.
				continue search
			}
		}

		// If we got here, it means that all of the bricks in the volume are
		// created! Continue on with the outer calling function.
		return
	}
}

// GenerateID will generate a unique identifier for this volume.
func (v *Volumes) GenerateID() string {
search:
	for {
		b := make([]byte, 8)
		rand.Read(b)
		id := hex.EncodeToString(b)

		for _, volume := range *v {
			if volume.ID == id {
				continue search
			}
		}

		return id
	}
}

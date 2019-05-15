package engine

import (
	"crypto/rand"
	"encoding/hex"
	"time"
)

// Volume is a distributed block storage volume propagated by GlusterFS.
type Volume struct {
	ID     string  `json:"id"`
	Name   string  `json:"name"`   // The short (friendly) name for the volume
	Size   int     `json:"size"`   // Size of volume in MB (used for allocation)
	Bricks []Brick `json:"bricks"` // The different bricks for this volume
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

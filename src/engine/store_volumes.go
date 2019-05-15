package engine

import (
	"crypto/rand"
	"encoding/hex"
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

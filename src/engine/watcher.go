package engine

import (
	"log"
	"os"
	"time"

	"orbit.sh/engine/gluster"
)

// Watcher is a process responsible for watching the processes taking place in
// the engine. it also keeps track of the engine so it can perform operations on
// it.
type Watcher struct {
	engine *Engine
}

// NewWatcher will return a new instance of a watcher.
func NewWatcher(e *Engine) *Watcher {
	return &Watcher{
		engine: e,
	}
}

// Start will start to watch the engine and handle the required state updates
// required. This involves anything on a node that must match the state present
// in the engine, such as the swap space data, or the gluster storage volumes.
func (w *Watcher) Start() {
	for {
		// This operation is performed continuously. That means checking every half
		// a second will ensure a responsive state update in response to changes,
		// but will also prevent locking down an entire thread for the duration of
		// the process.
		time.Sleep(time.Millisecond * 500)

		// Perform the different checks.
		w.CreateBricks()
		w.MountBricks()
		w.MountVolumes()
	}
}

// MountBricks will ensure that if we're a node that houses a brick, that we
// mount that brick correctly.
func (w *Watcher) MountBricks() {

}

// MountVolumes will go through the state of the system and mount the bricks and
// volumes that it needs to.
func (w *Watcher) MountVolumes() {
	for _, v := range w.engine.Store.state.Volumes {
		// Ensure that the most basic paths exist for this.
		paths := v.Paths()
		os.MkdirAll(paths.Data, 0644)

		// Find the IP address to use for the mount point.
		var ip string
	search:
		for _, b := range v.Bricks {
			for _, n := range w.engine.Store.state.Nodes {
				if n.ID == b.NodeID {
					ip = n.Address.String()
					break search
				}
			}
		}

		// Perform the mount of the gluster volume if it's not already mounted.
		gluster.MountGluster(ip, v.ID, paths.Data)
	}
}

// CreateBricks handles checking of the volume state.
func (w *Watcher) CreateBricks() {
	// Check if we need to create a brick.
	for _, v := range w.engine.Store.state.Volumes {
		for _, b := range v.Bricks {
			// The volume hasn't been created and the volume needs to be created for
			// this node. Perform this creation operation from the gluster package and
			// then update the store with the created state after this has taken
			// place.
			if !b.Created && b.NodeID == w.engine.Store.ID {
				paths := v.Paths()

				// Create the container directory.
				if err := os.MkdirAll(paths.Container, 0644); err != nil {
					log.Printf("[ERR] watcher: Could not create volume container directory %s: %s", paths.Container, err)
				}

				// Create the raw storage brick.
				if err := gluster.Fallocate(paths.Raw, v.Size); err != nil {
					log.Printf("[ERR] watcher: Could not create raw file %s: %s", paths.Raw, err)
				}

				// Make the raw storage brick into a filesystem.
				if err := gluster.MakeFS("xfs", paths.Raw); err != nil {
					log.Printf("[ERR] watcher: Could not convert %s to xfs filesystem: %s", paths.Raw, err)
				}

				// Create the volume directory.
				if err := os.MkdirAll(paths.Volume, 0644); err != nil {
					log.Printf("[ERR] watcher: Could not create volume directory %s: %s", paths.Volume, err)
				}

				// Mount the raw storage brick to the volume directory.
				if err := gluster.Mount(paths.Raw, paths.Volume); err != nil {
					log.Printf("[ERR] watcher: Could not mount %s to %s: %s", paths.Raw, paths.Volume, err)
				}

				// Create the brick directory.
				if err := os.MkdirAll(paths.Brick, 0644); err != nil {
					log.Printf("[ERR] watcher: Could not create volume brick directory %s: %s", paths.Brick, err)
				}

				// Create the data directory.
				if err := os.MkdirAll(paths.Data, 0644); err != nil {
					log.Printf("[ERR] watcher: Could not create volume data directory %s: %s", paths.Data, err)
				}

				// Apply the created marker.
				cmd := command{
					Op:     opUpdateVolumeBrick,
					Volume: Volume{ID: v.ID},
					Brick: Brick{
						NodeID:  b.NodeID,
						Created: true,
					},
				}

				if err := cmd.Apply(w.engine.Store); err != nil {
					log.Printf("[ERR] watcher: Could not set the volume %s brick at %s to have created: true: %s", v.ID, b.NodeID, err)
				}
			}
		}
	}
}

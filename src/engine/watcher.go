package engine

import "time"

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
		w.volumes()
	}
}

// volumes handles checking of the volume state.
func (w *Watcher) volumes() {
	// Check if we need to create a brick.
	for _, v := range w.engine.Store.state.Volumes {
		for _, b := range v.Bricks {
			// The volume hasn't been created and the volume needs to be created for
			// this node. Perform this creation operation from the gluster package and
			// then update the store with the created state after this has taken
			// place.
			if !b.Created && b.NodeID == w.engine.Store.ID {
				// TODO: Perform the creation.
			}
		}
	}
}

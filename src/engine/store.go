package engine

import (
	"net"
	"sync"
	"time"

	"github.com/hashicorp/raft"
)

// Store is a replicated state machine, where all changes are made via Raft
// distributed consensus.
type Store struct {
	engine *Engine // The engine instance that the store is tied to

	AdvertiseAddr net.IP
	RaftPort      int
	SerfPort      int
	WANSerfPort   int

	RetainSnapshotCount int
	RaftTimeout         time.Duration
	RaftMaxPool         int

	mu    sync.Mutex
	state *State
	raft  *raft.Raft // Primary consensus mechanism
}

// NewStore returns a new instance of the store.
func NewStore(e *Engine) *Store {
	return &Store{
		engine: e,

		RaftPort:    6502,
		SerfPort:    6503,
		WANSerfPort: 6504,

		RetainSnapshotCount: 2,
		RaftTimeout:         10 * time.Second,
		RaftMaxPool:         7,

		state: &State{},
	}
}

package engine

import (
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/hashicorp/raft"
	raftboltdb "github.com/hashicorp/raft-boltdb"
	"github.com/pkg/errors"
)

// Store is a replicated state machine, where all changes are made via Raft
// distributed consensus.
type Store struct {
	engine *Engine // The engine instance that the store is tied to

	AdvertiseAddr net.IP
	RaftPort      int
	SerfPort      int
	WANSerfPort   int
	ID            string

	RetainSnapshotCount int
	RaftTimeout         time.Duration
	RaftMaxPool         int

	mu    sync.RWMutex
	state *StoreState
	raft  *raft.Raft // Primary consensus mechanism

	started sync.WaitGroup
}

// NewStore returns a new instance of the store.
func NewStore(e *Engine) *Store {
	s := &Store{
		engine: e,

		RaftPort:    6502,
		SerfPort:    6503,
		WANSerfPort: 6504,

		RetainSnapshotCount: 2,
		RaftTimeout:         10 * time.Second,
		RaftMaxPool:         7,

		state: &StoreState{},
	}

	s.started.Add(1)

	return s
}

// Started is when the store has been started.
func (s *Store) Started() <-chan struct{} {
	ch := make(chan struct{})

	go func() {
		s.started.Wait()
		close(ch)
	}()

	return ch
}

// Open will open an instance of the store.
//
// This will always return an error, and will otherwise not return until an
// error occurs. For the purposes of the engine, this should be used in a
// non-blocking context.
func (s *Store) Open() error {
	// Ensure that we have an advertise address and that it's valid.
	if s.AdvertiseAddr == nil {
		return fmt.Errorf("invalid advertise address")
	}

	// Generate node ID if one does not exist.
	if s.ID == "" {
		s.ID = s.state.Nodes.GenerateNodeID()
		s.engine.writeConfig() // Ensure we keep the ID
	}

	// Set up raft configuration.
	raftConfig := raft.DefaultConfig()
	raftConfig.LocalID = raft.ServerID(s.ID)

	// Set up raft communication.
	addr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d", s.AdvertiseAddr, s.RaftPort))
	if err != nil {
		return errors.Wrap(err, "could not resolve tcp address")
	}
	transport, err := raft.NewTCPTransport(fmt.Sprintf("%s", addr), addr, s.RaftMaxPool, s.RaftTimeout, os.Stderr)
	if err != nil {
		return errors.Wrap(err, "could not create tcp transport")
	}

	// Create the store instances.
	var (
		snapshotStore raft.SnapshotStore
		logStore      raft.LogStore
		stableStore   raft.StableStore
	)

	// Instantiate the store instances.
	{
		// The log store.
		logDB, err := raftboltdb.NewBoltStore(filepath.Join(s.engine.DataPath, "raft", "log.db"))
		if err != nil {
			return errors.Wrap(err, "could not create log store")
		}
		logStore = logDB

		// The stable store.
		stableDB, err := raftboltdb.NewBoltStore(filepath.Join(s.engine.DataPath, "raft", "stable.db"))
		if err != nil {
			return errors.Wrap(err, "could not create stable store")
		}
		stableStore = stableDB

		// The snapshot store.
		snapshotDB, err := raft.NewFileSnapshotStore(
			filepath.Join(s.engine.DataPath, "raft", "snapshots"),
			s.RetainSnapshotCount,
			os.Stderr,
		)
		if err != nil {
			return errors.Wrap(err, "could not create snapshot store")
		}
		snapshotStore = snapshotDB
	}

	// Instantiate raft systems.
	ra, err := raft.NewRaft(raftConfig, (*fsm)(s), logStore, stableStore, snapshotStore, transport)
	if err != nil {
		return errors.Wrap(err, "could not instantiate raft")
	}
	s.raft = ra

	s.started.Done()
	log.Printf("[INFO] store: Opened at %s with ID %s", addr.String(), s.ID)
	select {}
}

// Bootstrap will actually start the store if it's the only node. This will only
// work if the store is not open or joined to another node.
func (s *Store) Bootstrap() error {
	if s.engine.Status == StatusRunning {
		err := fmt.Errorf("Cannot bootstrap a store that is already bootstrapped")
		log.Printf("[ERR] store: %s", err)
		return err
	}

	bootstrapConfig := raft.Configuration{
		Servers: []raft.Server{
			{
				ID:      raft.ServerID(s.ID),
				Address: raft.ServerAddress(fmt.Sprintf("%s:%d", s.AdvertiseAddr, s.RaftPort)),
			},
		},
	}
	s.raft.BootstrapCluster(bootstrapConfig)

	return nil
}

// CurrentNode is a helper method that returns a store state node object from
// both the current state of the store and the engine. The purpose of this is to
// make the &command{} to apply an easier process.
func (s *Store) CurrentNode() *Node {
	return &Node{
		ID:          s.ID,
		Address:     s.AdvertiseAddr,
		RPCPort:     s.engine.RPCServer.Port,
		RaftPort:    s.RaftPort,
		SerfPort:    s.SerfPort,
		WANSerfPort: s.WANSerfPort,
	}
}

// Join will join a node to this store instance. The node must be ready to
// respond to raft communications at that address (that means that the node must
// have a store instance running).
func (s *Store) Join(nodeID string, addr net.TCPAddr) error {
	log.Printf("[INFO] store: Received join request for node %s at %s", nodeID, addr.String())

	if s.raft.State() != raft.Leader {
		return fmt.Errorf("not leader")
		// TODO: Make this use forwardJoin
	}

	configFuture := s.raft.GetConfiguration()
	if err := configFuture.Error(); err != nil {
		log.Printf("[ERR] store: Failed to get raft configuration")
		return err
	}

	parsedAddr := raft.ServerAddress(fmt.Sprintf("%s", addr.String()))
	addVoterFuture := s.raft.AddVoter(raft.ServerID(nodeID), parsedAddr, 0, 0)
	if err := addVoterFuture.Error(); err != nil {
		log.Printf("[ERR] store: Could not add a voter to the cluster")
		return err
	}

	log.Printf("[INFO] store: Node %s at %s has joined successfully", nodeID, addr.String())
	return nil
}

// forwardJoin will forward a join cluster request to the leader of the cluster.
func forwardJoin(nodeID string, addr net.TCPAddr) error {
	// TODO: Implement the forward join method.
	return nil
}

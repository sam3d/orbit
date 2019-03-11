package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"

	"github.com/hashicorp/raft"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"orbit.sh/engine/proto"
)

type fsm Store

type op uint16

const (
	opEmpty op = iota

	opNewUser
	opRemoveUser

	opNewNode
)

type command struct {
	Op op `json:"op"`

	User User `json:"user,omitempty"`
	Node Node `json:"node,omitempty"`
}

// Apply is a helper proxy method that will apply the command to a raft instance
// in the store using it's "Apply" method. This is also the part of the process
// that is responsible for leader forwarding.
func (c *command) Apply(s *Store) error {
	b, err := json.Marshal(c)
	if err != nil {
		return err
	}

	// Ensure that we're the leader, and if not, forward the request to the
	// leader. This is a bit of a hacky implementation, but hopefully for now it
	// works.
	if s.raft.State() != raft.Leader {
		return forwardApply(s, b)
	}

	f := s.raft.Apply(b, s.RaftTimeout)
	return f.Error()
}

// forwardApply will apply a command by forwarding it to the current leader. This
// ensures that requests propagate correctly.
func forwardApply(s *Store, b []byte) error {
	leaderAddr := s.engine.RPCServer.Leader()
	if leaderAddr == "" {
		return fmt.Errorf("could not retrieve leader address")
	}

	log.Printf("[INFO] store: Forwarding to %s", leaderAddr)

	// Prepare the GRPC connection.
	conn, err := grpc.Dial(leaderAddr, grpc.WithInsecure())
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("could not dial %s", leaderAddr))
	}
	defer conn.Close()
	client := proto.NewRPCClient(conn)

	// Make the request.
	res, err := client.Apply(context.Background(), &proto.ApplyRequest{Body: b})
	if err != nil {
		return errors.Wrap(err, "could not perform the remote apply request")
	}
	if res.Status == proto.Status_ERROR {
		return fmt.Errorf("error from the leader node that we forwarded the request to")
	}

	return nil
}

// Apply will apply an entry to the store.
func (f *fsm) Apply(l *raft.Log) interface{} {
	var c command
	if err := json.Unmarshal(l.Data, &c); err != nil {
		panic("failed to unmarshal command")
	}

	switch c.Op {
	// User operations
	case opNewUser:
		return f.applyNewUser(c.User)
	case opRemoveUser:
		return f.applyRemoveUser(c.User.ID)

		// Node operations
	case opNewNode:
		return f.applyNewNode(c.Node)
	}

	return nil
}

func (f *fsm) applyNewUser(u User) interface{} {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.state.Users = append(f.state.Users, u)
	return nil
}

func (f *fsm) applyRemoveUser(id string) interface{} {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.state.Users.Remove(id)
	return nil
}

func (f *fsm) applyNewNode(n Node) interface{} {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.state.Nodes = append(f.state.Nodes, n)
	return nil
}

// Snapshot is a method that a raft finite state machine requires to operate. It
// simply copies the data into an FSM snapshot.
func (f *fsm) Snapshot() (raft.FSMSnapshot, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	snapshot := &fsmSnapshot{}
	if err := copier.Copy(snapshot, f.state); err != nil {
		log.Println("[ERR] fsm: Could not copy data into snapshot")
		return nil, err
	}

	return snapshot, nil
}

// Restore will restore the store state back to a previous state.
func (f *fsm) Restore(rc io.ReadCloser) error {
	state := &StoreState{}

	if err := json.NewDecoder(rc).Decode(state); err != nil {
		log.Println("[ERR] fsm: Could not restore snapshot")
		return err
	}

	// Set the state from the snapshot. This does not require a mutex lock.
	f.state = state
	return nil
}

// fsmSnapshot is an instance of the fsm state that implements the required
// methods to function as a snapshot. This should be cloned from the fsm
// instance so as to avoid corrupting any of the data latent there.
type fsmSnapshot StoreState

// Persist will take the data from the snapshot method and marshal it into data
// that it can persist on the disk as a binary blob.
//
// The data is taken back on the fsm using the Restore method.
func (f *fsmSnapshot) Persist(sink raft.SnapshotSink) error {
	// This is encapsulated in a method so that we ensure that we cancel the sink
	// should an error occur at any point in the process and can also ensure that
	// no other methods get written. It's the cleanest way to implement this.
	err := func() error {
		// Encode the store data.
		b, err := json.Marshal(f)
		if err != nil {
			return err
		}

		// Write the data to the sink.
		if _, err := sink.Write(b); err != nil {
			return err
		}

		// Close the sink.
		return sink.Close()
	}()

	// If an error occurred at any point in the process, ensure that we cancel the
	// process.
	if err != nil {
		sink.Cancel()
	}

	return err
}

func (f *fsmSnapshot) Release() {}

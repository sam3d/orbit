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

func (f *fsm) Snapshot() (raft.FSMSnapshot, error) {
	return nil, nil
}

func (f *fsm) Restore(rc io.ReadCloser) error {
	return nil
}

type fsmSnapshot struct{}

func (f *fsmSnapshot) Persist(sink raft.SnapshotSink) error {
	return nil
}

func (f *fsmSnapshot) Release() {}

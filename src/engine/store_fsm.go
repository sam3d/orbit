package engine

import (
	"encoding/json"
	"io"

	"github.com/hashicorp/raft"
)

type fsm Store

type command struct {
	Op   string `json:"op"`
	User User   `json:"user,omitempty"`
}

// Apply is a helper proxy method that will apply the command to a raft instance
// in the store using it's "Apply" method.
func (c *command) Apply(s *Store) error {
	b, err := json.Marshal(c)
	if err != nil {
		return err
	}
	f := s.raft.Apply(b, s.RaftTimeout)
	return f.Error()
}

// Apply will apply an entry to the store.
func (f *fsm) Apply(l *raft.Log) interface{} {
	var c command
	if err := json.Unmarshal(l.Data, &c); err != nil {
		panic("failed to unmarshal command")
	}

	switch c.Op {
	case "User.New":
		return f.applyNewUser(c.User)
	case "User.Remove":
		return f.applyRemoveUser(c.User.ID)
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

package engine

import (
	"encoding/json"
	"io"

	"github.com/hashicorp/raft"
)

type fsm Store

type command struct {
	Namespace string `json:"namespace"`
	User      User   `json:"user,omitempty"`
}

// Apply will apply an entry to the store.
func (f *fsm) Apply(l *raft.Log) interface{} {
	var c command
	if err := json.Unmarshal(l.Data, &c); err != nil {
		panic("failed to unmarshal command")
	}

	switch c.Namespace {
	case "User.New":
		return f.applyNewUser(c.User)
	}

	return nil
}

func (f *fsm) applyNewUser(u User) interface{} {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.state.Users = append(f.state.Users, u)
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

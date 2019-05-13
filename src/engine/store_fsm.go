package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"

	"github.com/hashicorp/raft"
	"github.com/jinzhu/copier"
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
	opNewSession
	opRevokeSession
	opRevokeAllSessions

	opNewNode
	opUpdateNode

	opNewNamespace

	opNewRouter
	opUpdateRouter

	opNewCertificate
	opUpdateCertificate
)

type command struct {
	Op op `json:"op"`

	User        User        `json:"user,omitempty"`
	Session     Session     `json:"session,omitempty"`
	Node        Node        `json:"node,omitempty"`
	Router      Router      `json:"router,omitempty"`
	Certificate Certificate `json:"certificate,omitempty"`
	Namespace   Namespace   `json:"namespace,omitempty"`
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
	// User operations.
	case opNewUser:
		return f.applyNewUser(c.User)
	case opRemoveUser:
		return f.applyRemoveUser(c.User.ID)
	case opNewSession:
		return f.applyNewSession(c.User.ID, c.Session)
	case opRevokeSession:
		return f.applyRevokeSession(c.Session.Token)
	case opRevokeAllSessions:
		return f.applyRevokeAllSessions(c.User.ID)

	// Node operations.
	case opNewNode:
		return f.applyNewNode(c.Node)
	case opUpdateNode:
		return f.applyUpdateNode(c.Node)

	// Namespace operations.
	case opNewNamespace:
		return f.applyNewNamespace(c.Namespace)

	// Router and certificate operations.
	case opNewRouter:
		return f.applyNewRouter(c.Router)
	case opUpdateRouter:
		return f.applyUpdateRouter(c.Router)

	case opNewCertificate:
		return f.applyNewCertificate(c.Certificate)
	case opUpdateCertificate:
		return f.applyUpdateCertificate(c.Certificate)
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

func (f *fsm) applyNewSession(id string, session Session) interface{} {
	f.mu.Lock()
	defer f.mu.Unlock()

	// Find the user and add the session to their list of sessions.
	for i, u := range f.state.Users {
		if u.ID == id {
			f.state.Users[i].Sessions = append(u.Sessions, session)
			break
		}
	}

	return nil
}

func (f *fsm) applyRevokeSession(token string) interface{} {
	f.mu.Lock()
	defer f.mu.Unlock()

	// Search through the users and sessions for that token, and if found, delete
	// it from that user.
search:
	for i, u := range f.state.Users {
		for j, s := range u.Sessions {
			if s.Token == token {
				// It was found, remove it and stop the loops.
				f.state.Users[i].Sessions = append(u.Sessions[:j], u.Sessions[j+1:]...)
				break search
			}
		}
	}

	return nil
}

func (f *fsm) applyRevokeAllSessions(userID string) interface{} {
	f.mu.Lock()
	defer f.mu.Unlock()

	// Find the user in question and set the sessions to an empty slice.
	for i, u := range f.state.Users {
		if u.ID == userID {
			f.state.Users[i].Sessions = []Session{}
			break
		}
	}

	return nil
}

func (f *fsm) applyNewNode(n Node) interface{} {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.state.Nodes = append(f.state.Nodes, n)
	return nil
}

func (f *fsm) applyUpdateNode(n Node) interface{} {
	f.mu.Lock()
	defer f.mu.Unlock()

	// Find the existing node so we can keep the details from it, and then remove
	// it from the store so that we can add it again.
	var currentNode Node
	var foundNode bool
	for i, node := range f.state.Nodes {
		if node.ID == n.ID {
			f.state.Nodes = append(f.state.Nodes[:i], f.state.Nodes[i+1:]...)
			currentNode = node
			foundNode = true
			break
		}
	}
	if !foundNode {
		return nil
	}

	// Handle all of the properties that could get updated.
	currentNode.Roles = n.Roles

	if n.SwapSize != -1 {
		currentNode.SwapSize = n.SwapSize
	}
	if n.Swappiness != -1 {
		currentNode.Swappiness = n.Swappiness
	}

	// Re-create the node in the store.
	f.state.Nodes = append(f.state.Nodes, currentNode)
	return nil
}

func (f *fsm) applyNewNamespace(n Namespace) interface{} {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.state.Namespaces = append(f.state.Namespaces, n)
	return nil
}

func (f *fsm) applyNewRouter(r Router) interface{} {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.state.Routers = append(f.state.Routers, r)
	return nil
}

func (f *fsm) applyNewCertificate(c Certificate) interface{} {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.state.Certificates = append(f.state.Certificates, c)
	return nil
}

func (f *fsm) applyUpdateCertificate(c Certificate) interface{} {
	f.mu.Lock()
	defer f.mu.Unlock()

	// Find the existing certificate so we can keep the details from it, and then
	// remove it from the store so that we can add it again.
	var currentCertificate Certificate
	var foundCertificate bool
	for i, cert := range f.state.Certificates {
		if cert.ID == c.ID {
			f.state.Certificates = append(f.state.Certificates[:i], f.state.Certificates[i+1:]...)
			currentCertificate = cert
			foundCertificate = true
			break
		}
	}
	if !foundCertificate {
		return nil
	}

	// We need to completely overwrite all of the challenges on a certificate no
	// matter what happens when we update a certificate. This ensures that we can
	// always clear out the pending challenges and add new ones when we perform
	// the update.
	currentCertificate.Challenges = c.Challenges

	// Update other details normally (if they are not equal to their null
	// counterparts).
	if len(c.FullChain) > 0 {
		currentCertificate.FullChain = c.FullChain
	}
	if len(c.PrivateKey) > 0 {
		currentCertificate.PrivateKey = c.PrivateKey
	}

	// Apply the new certificate.
	f.state.Certificates = append(f.state.Certificates, currentCertificate)
	return nil
}

func (f *fsm) applyUpdateRouter(r Router) interface{} {
	f.mu.Lock()
	defer f.mu.Unlock()

	// Find the existing router so we can keep the details from it, and then
	// remove it from the store so that we can add it again.
	currentRouter := func() *Router {
		for i, router := range f.state.Routers {
			if router.ID == r.ID {
				f.state.Routers = append(f.state.Routers[:i], f.state.Routers[i+1:]...)
				return &router
			}
		}
		return nil
	}()

	// Update the router object if the properties have been specified.
	if r.CertificateID != "" {
		currentRouter.CertificateID = r.CertificateID
	}
	if r.Domain != "" {
		currentRouter.Domain = r.Domain
	}
	if r.NamespaceID != "" {
		currentRouter.NamespaceID = r.NamespaceID
	}
	if r.AppID != "" {
		currentRouter.AppID = r.AppID
	}

	// Re-create the router object.
	f.state.Routers = append(f.state.Routers, *currentRouter)
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

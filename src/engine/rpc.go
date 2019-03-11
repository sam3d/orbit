package engine

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
	"orbit.sh/engine/proto"
)

// RPCServer is a remote server that hosts intra-node communications.
type RPCServer struct {
	engine  *Engine      // Keep track of the Orbit Engine instance that created it.
	server  *grpc.Server // The primary gRPC server instance
	Port    int
	started sync.WaitGroup
}

// NewRPCServer returns a new instance of the RPC Server.
func NewRPCServer(e *Engine) *RPCServer {
	s := &RPCServer{
		engine: e,
		server: grpc.NewServer(),
	}
	s.started.Add(1)
	return s
}

// Started will return a signal channel that closes when the server has started.
func (s *RPCServer) Started() <-chan struct{} {
	ch := make(chan struct{})

	go func() {
		s.started.Wait()
		close(ch)
	}()

	return ch
}

// Start will start the RPC server. It will only return if there is an error,
// otherwise it will hang forever.
func (s *RPCServer) Start() error {
	// Register the RPC server. This uses a GRPC package that is auto generated.
	proto.RegisterRPCServer(s.server, s)

	// Create the TCP listener.
	listenAddr := fmt.Sprintf(":%d", s.Port)
	listener, err := net.Listen("tcp", listenAddr)
	if err != nil {
		return errors.Wrap(err, "could not bind tcp listener")
	}

	// Start the RPC server asynchronously.
	errCh := make(chan error)
	go func() { errCh <- s.server.Serve(listener) }()

	log.Printf("[INFO] rpc: Listening on port %v", s.Port)
	s.started.Done()
	return <-errCh
}

// Join handle receiving an RPC to join the server.
func (s *RPCServer) Join(ctx context.Context, in *proto.JoinRequest) (*proto.JoinResponse, error) {
	engine := s.engine
	store := engine.Store

	// Begin constructing the response.
	res := &proto.JoinResponse{
		RaftPort:    uint32(store.RaftPort),
		SerfPort:    uint32(store.SerfPort),
		WanSerfPort: uint32(store.WANSerfPort),
		Status:      proto.Status_OK,
	}

	// Ensure that the server is ready to receive connections.
	if engine.Status != StatusRunning {
		res.Status = proto.Status_ERROR
		return res, nil
	}

	// Ensure that the join token is valid.
	if in.JoinToken != "" {
		res.Status = proto.Status_UNAUTHORIZED
		return res, nil
	}

	// Retrieve the client IP from the context.
	//
	// This is important, as this is how the node wishing to join the cluster has
	// been able to get hold of us. That means that we need to advise that their
	// advertise address is this one, as this is the one that they can reach us
	// from.
	p, _ := peer.FromContext(ctx)
	addr, _ := net.ResolveTCPAddr("tcp", p.Addr.String())
	ip := addr.IP.String()
	res.AdvertiseAddr = ip

	// Generate an ID for the node.
	id := store.state.Nodes.GenerateNodeID()
	res.Id = id

	return res, nil
}

// ConfirmJoin handles a node after it has been given the required data
// from the store. This will actually perform the join operation and create the
// node.
func (s *RPCServer) ConfirmJoin(ctx context.Context, in *proto.ConfirmJoinRequest) (*proto.StatusResponse, error) {
	engine := s.engine
	store := engine.Store

	// Construct the response.
	res := &proto.StatusResponse{
		Status: proto.Status_OK,
	}

	// Ensure we have a valid join token.
	if in.JoinToken != "" {
		res.Status = proto.Status_UNAUTHORIZED
		return res, nil
	}

	// Perform the join operation.
	addr, _ := net.ResolveTCPAddr("tcp", in.RaftAddr)
	if err := store.Join(in.Id, *addr); err != nil {
		log.Printf("[ERR] store: Could not join %s to the store", in.RaftAddr)
		res.Status = proto.Status_ERROR
		return res, nil
	}

	return res, nil
}

// Apply is called when a node that isn't a leader forwards an fsm apply blob to
// us. That means that we're the leader of the cluster, so go us!
func (s *RPCServer) Apply(ctx context.Context, in *proto.ApplyRequest) (*proto.StatusResponse, error) {
	res := &proto.StatusResponse{
		Status: proto.Status_OK,
	}

	f := s.engine.Store.raft.Apply(in.Body, s.engine.Store.RaftTimeout)
	if err := f.Error(); err != nil {
		log.Printf("[ERR] store: %s", err)
		res.Status = proto.Status_ERROR
	}

	return res, nil
}

// ForwardJoin is the method that handles us receiving a join request forwarded
// to us from another node. This request is already authenticated, and it means
// that we are the leader o the cluster, so this simply has to be applied.
func (s *RPCServer) ForwardJoin(ctx context.Context, in *proto.ForwardJoinRequest) (*proto.StatusResponse, error) {
	log.Printf("[INFO] rpc: Received forwarded join request")

	res := &proto.StatusResponse{
		Status: proto.Status_OK,
	}

	addr, err := net.ResolveTCPAddr("tcp", in.Address)
	if err != nil {
		log.Printf("[ERR] rpc: Received forwarded request but can't parse TCP address: %v", err)
		res.Status = proto.Status_ERROR
		return res, nil
	}

	err = s.engine.Store.Join(in.NodeId, *addr)
	if err != nil {
		log.Printf("[ERR] rpc: Cannot perform store join operation: %v", err)
		res.Status = proto.Status_ERROR
		return res, nil
	}

	return res, nil
}

// Leader gets the RPC address of the leader of the cluster.
func (s *RPCServer) Leader() string {
	opTimeout := time.Second * 20             // Length of time for operation timeouts.
	opRetryInterval := time.Millisecond * 200 // Retry operations 5 times a second.

	// Add a timeout handler to ensure that there is a raft leader.
	for start := time.Now(); ; {
		if s.engine.Store.raft.Leader() != "" {
			break
		}
		if time.Since(start) > opTimeout {
			log.Printf("[ERR] rpc: could not get leader status")
			return ""
		}
		time.Sleep(opRetryInterval)
	}

	// Get the raft address of the leader.
	rawRaftAddr := string(s.engine.Store.raft.Leader())
	raftAddr, err := net.ResolveTCPAddr("tcp", rawRaftAddr)
	if err != nil {
		log.Printf("[ERR] rpc: Could not resolve TCP address of leader")
		return ""
	}
	ip := raftAddr.IP.String()

	// Look up the required port from the IP address in the store state node list.
	var rpcPort int
found:
	for start := time.Now(); ; {
		if time.Since(start) > opTimeout {
			log.Printf("[ERR] rpc: Our copy of the node list never got updated")
			return ""
		}

		for _, node := range s.engine.Store.state.Nodes {
			if node.Address.String() == ip {
				rpcPort = node.RPCPort
				break found
			}
		}

		time.Sleep(opRetryInterval)
	}
	if rpcPort == 0 {
		log.Printf("[ERR] rpc: Could not load the leader port in the store state node list")
		return ""
	}

	// Great, we found it out. Now let's return it.
	addr := fmt.Sprintf("%s:%d", ip, rpcPort)
	return addr
}

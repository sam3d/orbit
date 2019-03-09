package engine

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"

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
	proto.RegisterClusterServer(s.server, s)

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

// ClusterJoin handle receiving an RPC to join the server.
func (s *RPCServer) ClusterJoin(ctx context.Context, in *proto.ClusterJoinRequest) (*proto.ClusterJoinResponse, error) {
	engine := s.engine
	store := engine.Store

	// Begin constructing the response.
	res := &proto.ClusterJoinResponse{
		RaftPort:    uint32(store.RaftPort),
		SerfPort:    uint32(store.SerfPort),
		WanSerfPort: uint32(store.WANSerfPort),
		JoinStatus:  proto.ClusterStatus_OK,
	}

	// Ensure that the server is ready to receive connections.
	if engine.Status != StatusRunning {
		res.JoinStatus = proto.ClusterStatus_ERROR
		return res, nil
	}

	// Ensure that the join token is valid.
	if in.JoinToken != "" {
		res.JoinStatus = proto.ClusterStatus_UNAUTHORIZED
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

// ClusterJoinConfirm handles a node after it has been given the required data
// from the store. This will actually perform the join operation and create the
// node.
func (s *RPCServer) ClusterJoinConfirm(ctx context.Context, in *proto.ClusterJoinConfirmRequest) (*proto.ClusterJoinConfirmResponse, error) {
	engine := s.engine
	store := engine.Store

	// Construct the response.
	res := &proto.ClusterJoinConfirmResponse{
		ConfirmStatus: proto.ClusterStatus_OK,
	}

	// Ensure we have a valid join token.
	if in.JoinToken != "" {
		res.ConfirmStatus = proto.ClusterStatus_UNAUTHORIZED
		return res, nil
	}

	// Perform the join operation.
	addr, _ := net.ResolveTCPAddr("tcp", in.RaftAddr)
	if err := store.Join(in.Id, *addr); err != nil {
		log.Printf("[ERR] store: Could not join %s to the store", in.RaftAddr)
		res.ConfirmStatus = proto.ClusterStatus_ERROR
		return res, nil
	}

	return res, nil
}
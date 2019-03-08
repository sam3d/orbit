package engine

import "sync"

// RPCServer is a remote server that hosts intra-node communications.
type RPCServer struct {
	engine  *Engine // Keep track of the Orbit Engine instance that created it.
	Port    int
	started sync.WaitGroup
}

// NewRPCServer returns a new instance of the RPC Server.
func NewRPCServer(e *Engine) *RPCServer {
	s := &RPCServer{
		engine: e,
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
	s.started.Done()
	select {}
}

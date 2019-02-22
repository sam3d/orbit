// Package api provides an HTTP interface to the engine operations that exposes
// both on a port and on a UNIX socket. The CLI will, by default, dial into the
// UNIX socket that is exposed and consume the HTTP API.
package api

import (
	"fmt"
	"log"
	"sync"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.ReleaseMode)
}

// Server is the root instance for the API server.
type Server struct {
	router *gin.Engine
	Port   int
	Socket string

	startedWg sync.WaitGroup
}

// New returns a new API server instance.
func New() *Server {
	s := &Server{
		router: gin.New(),
	}

	// We need to set the waitgroup at start so that if the user requests the
	// started channel, it waits until it has started. We need to remember to
	// release this extra waitgroup lock after the first addition, and add it
	// again when the API is stopped.
	s.startedWg.Add(1)

	return s
}

// Started returns a channel as to whether or not the api has started.
func (s *Server) Started() <-chan struct{} {
	ch := make(chan struct{})

	go func() {
		s.startedWg.Wait()
		close(ch)
	}()

	return ch
}

// Start will start the server. It will *always* return an error from either the
// UNIX socket listener or the TCP listener (depending on which one errors
// first).
func (s *Server) Start() error {
	s.routes()              // Register the routes
	err := make(chan error) // Handle errors from socket and TCP

	s.startedWg.Add(2)
	s.startedWg.Done() // Clear out the initial waitgroup

	// Listen for UNIX socket requests.
	go func() {
		if s.Socket == "" {
			log.Println("[WARN] api: not listening for socket requests")
			s.startedWg.Done()
			return
		}

		log.Printf("[INFO] api: listening on socket %v", s.Socket)
		s.startedWg.Done()
		err <- s.router.RunUnix(s.Socket)
	}()

	// Listen for standard TCP requests.
	go func() {
		if s.Port == -1 {
			log.Println("[INFO] api: not listening for TCP requests")
			s.startedWg.Done()
			return
		}

		if s.Port < 0 || s.Port > 65535 {
			err <- fmt.Errorf("[ERR] api: port %d is out of range", s.Port)
			s.startedWg.Done()
			return
		}

		log.Printf("[WARN] api: listening on port %v", s.Port)
		s.startedWg.Done()
		bindAddr := fmt.Sprintf(":%d", s.Port)
		err <- s.router.Run(bindAddr)
	}()

	return <-err
}

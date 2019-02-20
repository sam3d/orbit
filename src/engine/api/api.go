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

	started chan struct{}
}

// New returns a new API server instance.
func New() *Server {
	return &Server{
		router:  gin.New(),
		started: make(chan struct{}),
	}
}

// Started returns a channel as to whether or not the api has started.
func (s *Server) Started() <-chan struct{} {
	return s.started
}

// Start will start the server. It will *always* return an error from either the
// UNIX socket listener or the TCP listener (depending on which one errors
// first).
func (s *Server) Start() error {
	s.routes()              // Register the routes
	err := make(chan error) // Handle errors from socket and TCP

	// Keep track of processes so we know when to start.
	var wg sync.WaitGroup
	wg.Add(2)

	// Listen for UNIX socket requests.
	go func() {
		if s.Socket == "" {
			log.Println("[WARN] api: not listening for socket requests")
			wg.Done()
			return
		}

		log.Printf("[INFO] api: listening on socket %v", s.Socket)
		wg.Done()
		err <- s.router.RunUnix(s.Socket)
	}()

	// Listen for standard TCP requests.
	go func() {
		if s.Port == -1 {
			log.Println("[INFO] api: not listening for TCP requests")
			wg.Done()
			return
		}

		if s.Port < 0 || s.Port > 65535 {
			err <- fmt.Errorf("[ERR] api: port %d is out of range", s.Port)
			wg.Done()
			return
		}

		log.Printf("[WARN] api: listening on port %v", s.Port)
		wg.Done()
		bindAddr := fmt.Sprintf(":%d", s.Port)
		err <- s.router.Run(bindAddr)
	}()

	// Check for started status.
	go func() {
		wg.Wait()
		close(s.started)
	}()

	return <-err
}

// Package api provides an HTTP interface to the engine operations that exposes
// both on a port and on a UNIX socket. The CLI will, by default, dial into the
// UNIX socket that is exposed and consume the HTTP API.
package api

import (
	"fmt"
	"log"

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
}

// New returns a new API server instance.
func New() *Server {
	return &Server{
		router: gin.Default(),
	}
}

// Start will start the server. It will *always* return an error from either the
// UNIX socket listener or the TCP listener (depending on which one errors
// first).
func (s *Server) Start() error {
	s.routes()              // Register the routes
	err := make(chan error) // Handle errors from socket and TCP

	// Listen for UNIX socket requests.
	go func() {
		if s.Socket == "" {
			log.Println("[WARN] api: not listening for socket requests")
			return
		}

		err <- s.router.RunUnix(s.Socket)
	}()

	// Listen for standard TCP requests.
	go func() {
		if s.Port == -1 {
			log.Println("[INFO] api: not listening for TCP requests")
			return
		}

		if s.Port < 0 || s.Port > 65535 {
			err <- fmt.Errorf("port %d is out of range", s.Port)
			return
		}

		log.Printf("[WARN] api: listening on port %v", s.Port)
		bindAddr := fmt.Sprintf(":%d", s.Port)
		err <- s.router.Run(bindAddr)
	}()

	return <-err
}

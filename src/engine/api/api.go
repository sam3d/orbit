// Package api provides an HTTP interface to the engine operations that exposes
// both on a port and on a UNIX socket. The CLI will, by default, dial into the
// UNIX socket that is exposed and consume the HTTP API.
package api

import (
	"fmt"

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

// NewServer returns a new API server instance.
func NewServer(socket string, port int) *Server {
	return &Server{
		router: gin.Default(),
		Port:   port,
		Socket: socket,
	}
}

// Start will start the server. This is simply a proxy for the internal engine
// that gin uses for routing. It will block the calling goroutine unless an
// error occurs in either the UNIX socket listener or the standard TCP address
// listener.
func (s *Server) Start() error {
	s.routes()              // Register the routes
	err := make(chan error) // Handle errors from socket and TCP

	// Listen for UNIX socket requests.
	go func() {
		if s.Socket != "" {
			err <- s.router.RunUnix(s.Socket)
		}
	}()

	// Listen for standard TCP requests.
	go func() {
		bindAddr := fmt.Sprintf(":%d", s.Port)
		err <- s.router.Run(bindAddr)
	}()

	return <-err
}

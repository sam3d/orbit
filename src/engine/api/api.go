package api

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
)

func init() {
	// gin.SetMode(gin.ReleaseMode)
}

// Server is the root instance for the API server.
type Server struct {
	router     *gin.Engine
	Port       int
	Host       string
	SocketPath string
	logger     *log.Logger
}

// New returns a new API server instance.
func New() *Server {
	return &Server{
		router:     gin.Default(),
		Port:       6501,
		Host:       "",
		SocketPath: "/var/run/orbit.sock",
	}
}

// Start will start the server. This is simply a proxy for the internal engine
// that gin uses for routing. It will block the calling goroutine unless an
// error occurs in either the UNIX socket listener or the standard TCP address
// listener.
func (s *Server) Start() error {
	s.routes()                // Register the routes
	errCh := make(chan error) // Handle errors from socket and TCP

	// Listen for UNIX socket requests.
	go func() {
		if s.SocketPath != "" {
			err := s.router.RunUnix(s.SocketPath)
			errCh <- err
		}
	}()

	// Listen for standard TCP requests.
	go func() {
		bindAddr := fmt.Sprintf("%s:%d", s.Host, s.Port)
		err := s.router.Run(bindAddr)
		errCh <- err
	}()

	return <-errCh
}

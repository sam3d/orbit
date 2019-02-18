package api

import (
	"fmt"
	"os"
	"strconv"

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
}

// New returns a new API server instance.
func New() (*Server, error) {
	// Parse socket path.
	socketPath, ok := os.LookupEnv("SOCKET_PATH")
	if !ok {
		socketPath = "/var/run/orbit.sock"
	}

	// Retrieve the port.
	port := 6501
	portEnv, ok := os.LookupEnv("PORT")
	if ok {
		// The port is being defined on the host, parse it.
		parsedPort, err := strconv.Atoi(portEnv)
		if err != nil {
			return nil, fmt.Errorf("PORT is not a valid integer")
		}
		port = parsedPort

		// Check for range errors.
		if port < 0 || port > 65535 {
			return nil, fmt.Errorf("PORT is out of range: must be between 0 and 65535")
		}
	}

	// Retrieve the host. By default, this can work as an empty string (which is
	// what the default value of os.Getenv will be).
	host := os.Getenv("HOST")

	return &Server{
		router:     gin.Default(),
		Port:       port,
		Host:       host,
		SocketPath: socketPath,
	}, nil
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

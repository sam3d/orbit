package engine

import (
	"fmt"
	"log"
	"sync"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.ReleaseMode)
}

// APIServer is the root instance for the API server.
type APIServer struct {
	engine *Engine

	Port   int
	Socket string

	router  *gin.Engine
	started sync.WaitGroup
}

// NewAPIServer returns a new API server instance.
func NewAPIServer(e *Engine) *APIServer {
	s := &APIServer{
		engine: e,
		router: gin.New(),
	}

	// We need to set the waitgroup at start so that if the user requests the
	// started channel, it waits until it has started. We need to remember to
	// release this extra waitgroup lock after the first addition, and add it
	// again when the API is stopped.
	s.started.Add(1)

	return s
}

// Started returns a channel as to whether or not the api has started.
func (s *APIServer) Started() <-chan struct{} {
	ch := make(chan struct{})

	go func() {
		s.started.Wait()
		close(ch)
	}()

	return ch
}

// Start will start the server. It will *always* return an error from either the
// UNIX socket listener or the TCP listener (depending on which one errors
// first).
func (s *APIServer) Start() error {
	s.handlers()              // Register the routes
	errCh := make(chan error) // Handle errors from socket and TCP

	s.started.Add(2)
	s.started.Done() // Clear out the initial waitgroup

	// Listen for UNIX socket requests.
	go func() {
		if s.Socket == "" {
			log.Println("[WARN] api: Not listening for socket requests")
			s.started.Done()
			return
		}

		log.Printf("[INFO] api: Listening on socket %s", s.Socket)
		s.started.Done()
		errCh <- s.router.RunUnix(s.Socket)
	}()

	// Listen for standard TCP requests.
	go func() {
		if s.Port == -1 {
			log.Println("[INFO] api: Not listening for TCP requests")
			s.started.Done()
			return
		}

		if s.Port < 0 || s.Port > 65535 {
			errCh <- fmt.Errorf("[ERR] api: Port %d is out of range", s.Port)
			s.started.Done()
			return
		}

		log.Printf("[WARN] api: Listening on port %d", s.Port)
		s.started.Done()
		bindAddr := fmt.Sprintf(":%d", s.Port)
		errCh <- s.router.Run(bindAddr)
	}()

	return <-errCh
}

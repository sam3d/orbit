package engine

import (
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
)

// RPCServer is the process that responds to requests from other agents.
type RPCServer struct {
	engine  *Engine // Keep track of the engine instance that created it
	router  *gin.Engine
	Port    int
	started sync.WaitGroup
}

// Started will close the channel once the process has started. This will start
// as a blocking process.
func (s *RPCServer) Started() <-chan struct{} {
	ch := make(chan struct{})
	go func() {
		s.started.Wait()
		close(ch)
	}()
	return ch
}

// NewRPCServer will return a new instance of the RPC server.
func NewRPCServer(e *Engine) *RPCServer {
	s := &RPCServer{
		engine: e,
		router: gin.New(),
	}
	s.started.Add(1)
	return s
}

// Start will start a new instance of the RPC server.
func (s *RPCServer) Start() error {
	s.handlers() // Register the route handlers

	log.Printf("[INFO] rpc: Listening on port %d", s.Port)
	s.started.Done()
	bindAddr := fmt.Sprintf(":%d", s.Port)
	return s.router.Run(bindAddr)
}

// handlers will register all of the handlers for the RPC server.
func (s *RPCServer) handlers() {
	r := s.router // Retrieve the router from the server

	// Log out all requests.
	r.Use(s.simpleLogger())

	//
	// Register all other routes.
	//

	r.GET("/", s.handleIndex())
	r.POST("/cluster/join", s.handleClusterJoin())
}

func (s *RPCServer) simpleLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Printf("[INFO] rpc: Received %s at %s", c.Request.Method, c.Request.URL)
		c.Next()
	}
}

func (s *RPCServer) handleIndex() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.String(http.StatusOK, "RPC is operational and responding to requests.")
	}
}

func (s *RPCServer) handleClusterJoin() gin.HandlerFunc {
	engine := s.engine
	store := engine.Store

	return func(c *gin.Context) {
		// Parse join request.
		var body RPCJoinRequest
		if err := c.Bind(&body); err != nil {
			c.Status(http.StatusBadRequest)
			return
		}

		// Check join token.
		if body.JoinToken != "" {
			c.String(http.StatusUnauthorized, "Invalid join token.")
			return
		}

		// Generate node ID and retrieve advertise address.
		addr := c.ClientIP()
		id := store.state.Nodes.GenerateNodeID()

		c.JSON(http.StatusOK, &RPCJoinResponse{
			AdvertiseAddr: addr,
			ID:            id,
		})
	}
}

package engine

import (
	"fmt"
	"log"
	"net"
	"net/http"
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

// handlers registers all of the default routes for the API server. This is a
// separate method so that other routes can be added *after* the defaults but
// *before* the server is started.
func (s *APIServer) handlers() {
	r := s.router

	// Register middleware.
	r.Use(s.simpleLogger())

	//
	// Handle all of the routes.
	//

	r.GET("", func(c *gin.Context) {
		c.String(http.StatusOK, "Welcome to the Orbit Engine API.\nAll systems are operational.")
	})

	r.GET("/state", s.handleState())
	r.GET("/users", s.handleListUsers())

	{
		r := r.Group("/cluster")
		r.POST("/bootstrap", s.handleClusterBootstrap())
		r.POST("/join", s.handleClusterJoin())
	}

	{
		r := r.Group("/user")
		r.POST("", s.handleUserSignup())
		r.DELETE("/:id", s.handleUserRemove())
	}
}

func (s *APIServer) simpleLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Printf("[INFO] api: Received %s at %s", c.Request.Method, c.Request.URL)
		c.Next()
	}
}

func (s *APIServer) handleState() gin.HandlerFunc {
	type res struct {
		Status       Status      `json:"status"`
		StatusString string      `json:"status_string"`
		State        *StoreState `json:"state"`
	}

	return func(c *gin.Context) {
		c.JSON(http.StatusOK, &res{
			Status:       s.engine.Status,
			StatusString: fmt.Sprintf("%s", s.engine.Status),
			State:        s.engine.Store.state,
		})
	}
}

func (s *APIServer) handleClusterBootstrap() gin.HandlerFunc {
	engine := s.engine
	store := engine.Store
	var mu sync.Mutex

	type body struct {
		RawAdvertiseAddr string `form:"advertise_address" json:"advertise_address"`
		RPCPort          int    `form:"rpc_port" json:"rpc_port"`
		RaftPort         int    `form:"raft_port" json:"raft_port"`
		SerfPort         int    `form:"serf_port" json:"serf_port"`
		WANSerfPort      int    `form:"wan_serf_port" json:"wan_serf_port"`
	}

	return func(c *gin.Context) {
		// Don't allow anybody else to attempt to bootstrap during process.
		mu.Lock()
		defer mu.Unlock()

		// Ensure that the store can be bootstrapped.
		if engine.Status >= StatusReady {
			c.String(http.StatusConflict, "The engine is already running and cannot be bootstrapped.")
			return
		}

		// Bind the default settings from the body.
		body := body{
			RPCPort:     6501,
			RaftPort:    6502,
			SerfPort:    6503,
			WANSerfPort: 6504,
		}
		if err := c.Bind(&body); err != nil {
			c.String(http.StatusBadRequest, "Invalid request body fields.")
			return
		}

		// Validate and parse the provided IP address.
		var advertiseAddr net.IP
		if body.RawAdvertiseAddr != "" {
			ip := net.ParseIP(body.RawAdvertiseAddr)
			if ip == nil {
				c.String(http.StatusBadRequest, "Your provided advertise address is not valid.")
				return
			}
			advertiseAddr = ip
		} else {
			ip, err := getPublicIP()
			if err != nil {
				c.String(http.StatusBadRequest, "Could not automatically obtain public advertise address.")
			}
			advertiseAddr = ip
		}

		// Set all of the engine component properties.
		engine.RPCServer.Port = body.RPCPort
		store.AdvertiseAddr = advertiseAddr
		store.RaftPort = body.RaftPort
		store.SerfPort = body.SerfPort
		store.WANSerfPort = body.WANSerfPort

		// Attempt to open the store.
		errCh := make(chan error)
		go func() { errCh <- store.Open() }()
		select {
		case <-store.Started():
		case err := <-errCh:
			log.Printf("[ERR] store: %s", err)
			c.String(http.StatusInternalServerError, "Could not open the store instance to bootstrap.")
			return
		}

		// Attempt to bootstrap the store.
		if err := store.Bootstrap(); err != nil {
			c.String(http.StatusInternalServerError, "Could not bootstrap the store instance.")
			return
		}

		// Save the state and set the engine status.
		engine.Status = StatusRunning
		engine.writeConfig()

		// TODO: Add the node to the store_nodes list.

		c.JSON(http.StatusOK, engine.marshalConfig())
	}
}

func (s *APIServer) handleClusterJoin() gin.HandlerFunc {
	engine := s.engine
	store := engine.Store
	type body struct {
		// Local node options.
		RPCPort     int `form:"rpc_port" json:"rpc_port"`
		RaftPort    int `form:"raft_port" json:"raft_port"`
		SerfPort    int `form:"serf_port" json:"serf_port"`
		WANSerfPort int `form:"wan_serf_port" json:"wan_serf_port"`

		// Options for node to join.
		RawTargetAddr string `form:"target_address" json:"target_address"` // RPC address of target node
		JoinToken     string `form:"join_token" json:"join_token"`
	}

	return func(c *gin.Context) {
		// Bind the default settings from the body.
		body := body{
			RPCPort:     6501,
			RaftPort:    6502,
			SerfPort:    6503,
			WANSerfPort: 6504,
		}
		if err := c.Bind(&body); err != nil {
			c.String(http.StatusBadRequest, "Invalid form fields.")
			return
		}

		// Validate the target address.
		targetAddr, err := net.ResolveTCPAddr("tcp", body.RawTargetAddr)
		if err != nil || len(targetAddr.IP) == 0 || targetAddr.Port == 0 {
			c.String(http.StatusBadRequest, "Invalid TCP target address.")
			return
		}
		client := RPCClient{Addr: *targetAddr}

		// Make the join request.
		var data RPCJoinResponse
		resp, err := client.Post("/cluster/join", &RPCJoinRequest{JoinToken: body.JoinToken}, &data)
		if err != nil {
			if resp != nil && resp.StatusCode == http.StatusUnauthorized {
				c.String(http.StatusUnauthorized, "Invalid join token.")
			} else {
				log.Printf("[ERR] api: %v", err)
				c.String(http.StatusInternalServerError, "Could not join target node.")
			}
			return
		}

		// Set up local properties.
		store.RaftPort = body.RaftPort
		store.SerfPort = body.SerfPort
		store.WANSerfPort = body.WANSerfPort

		// Set up the remote join properties.
		store.AdvertiseAddr = net.ParseIP(data.AdvertiseAddr)
		localRaftAddr := fmt.Sprintf("%s:%s", store.AdvertiseAddr, store.RaftPort)

		c.JSON(http.StatusOK, &data)
	}
}

func (s *APIServer) handleUserSignup() gin.HandlerFunc {
	store := s.engine.Store

	type body struct {
		Name     string `form:"name" json:"name"`
		Password string `form:"password" json:"password"`
		Username string `form:"username" json:"username"`
		Email    string `form:"email" json:"email"`
	}

	return func(c *gin.Context) {
		var body body
		c.Bind(&body)

		newUser, err := store.state.Users.Generate(UserConfig{
			Name:     body.Name,
			Password: body.Password,
			Username: body.Username,
			Email:    body.Email,
		})
		if err != nil {
			switch err {
			case ErrEmailTaken:
				c.String(http.StatusConflict, "Sorry, that email address is already taken.")
			case ErrUsernameTaken:
				c.String(http.StatusConflict, "Sorry, that username is already taken.")
			case ErrMissingFields:
				c.String(http.StatusBadRequest, "You didn't supply all of the required fields.")
			default:
				c.AbortWithStatus(http.StatusBadRequest)
			}
			return
		}

		cmd := command{
			Op:   "User.New",
			User: *newUser,
		}

		if err := cmd.Apply(store); err != nil {
			c.String(http.StatusInternalServerError, "Could not create the new user. Ensure that all of the manager nodes are connected correctly.")
			return
		}

		c.String(http.StatusCreated, newUser.ID)
	}
}

func (s *APIServer) handleListUsers() gin.HandlerFunc {
	store := s.engine.Store

	type user struct {
		ID       string `json:"id"`
		Name     string `json:"name"`
		Username string `json:"username"`
		Email    string `json:"email"`
	}

	return func(c *gin.Context) {
		store.mu.RLock()
		defer store.mu.RUnlock()

		var users []user

		for _, u := range store.state.Users {
			users = append(users, user{
				ID:       u.ID,
				Name:     u.Name,
				Username: u.Username,
				Email:    u.Email,
			})
		}

		c.JSON(http.StatusOK, &users)
	}
}

func (s *APIServer) handleUserRemove() gin.HandlerFunc {
	store := s.engine.Store

	return func(c *gin.Context) {
		id := c.Param("id") // The ID of the user to remove

		if i, _ := store.state.Users.FindByID(id); i == -1 {
			c.String(http.StatusNotFound, "A user with that ID does not exist.")
			return
		}

		cmd := command{
			Op:   "User.Remove",
			User: User{ID: id},
		}

		if err := cmd.Apply(store); err != nil {
			c.String(http.StatusInternalServerError, "Could not remove that user.")
			return
		}

		c.String(http.StatusOK, "The user has been removed.")
	}
}

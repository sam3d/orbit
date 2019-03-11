package engine

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/hashicorp/raft"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"orbit.sh/engine/proto"
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
	r.GET("/nodes", s.handleListNodes())

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
		Status       Status `json:"status"`
		StatusString string `json:"status_string"`
	}

	return func(c *gin.Context) {
		c.JSON(http.StatusOK, &res{
			Status:       s.engine.Status,
			StatusString: fmt.Sprintf("%s", s.engine.Status),
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

		// Attempt to start the RPC server.
		go func() { errCh <- engine.RPCServer.Start() }()
		select {
		case <-engine.RPCServer.Started():
		case err := <-errCh:
			log.Printf("[ERR] store: %s", err)
			c.String(http.StatusInternalServerError, "Could not start the RPC server.")
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

		// Wait for us to become the leader of the store.
		select {
		case leader := <-store.raft.LeaderCh():
			if !leader {
				log.Printf("[ERR] store: we did not become the leader of the store")
				c.String(http.StatusInternalServerError, "There was an error establishing a leader for the cluster.")
				return
			}
		case <-time.After(time.Second * 10):
			log.Printf("[ERR] store: we never received leader information")
			c.String(http.StatusInternalServerError, "There was an error establishing a leader for the cluster.")
			return
		}

		// Prepare command to add this node's details to the store.
		cmd := command{
			Op:   opNewNode,
			Node: *store.CurrentNode(),
		}

		if err := cmd.Apply(store); err != nil {
			log.Printf("[ERR] store: %s", err)
			c.String(http.StatusInternalServerError, "Could not add this node to the list of nodes in the store, despite being joined to it successfully.")
			return
		}

		c.JSON(http.StatusOK, engine.marshalConfig())
	}
}

func (s *APIServer) handleClusterJoin() gin.HandlerFunc {
	engine := s.engine
	store := engine.Store

	var mu sync.Mutex

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
		// Prevent using this route twice.
		mu.Lock()
		defer mu.Unlock()

		if engine.Status >= StatusReady {
			c.String(http.StatusConflict, "The node is already part of a cluster.")
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
			c.String(http.StatusBadRequest, "Invalid form fields.")
			return
		}

		// Validate the target address.
		targetAddr, err := net.ResolveTCPAddr("tcp", body.RawTargetAddr)
		if err != nil || len(targetAddr.IP) == 0 || targetAddr.Port == 0 {
			c.String(http.StatusBadRequest, "Invalid TCP target address.")
			return
		}

		// Create the client for connecting to the target node.
		conn, err := grpc.Dial(targetAddr.String(), grpc.WithInsecure())
		if err != nil {
			c.String(http.StatusBadRequest, "Could not establish a connection to %s.", targetAddr)
			return
		}
		defer conn.Close()
		client := proto.NewRPCClient(conn)

		// Actually make the join request.
		joinRes, err := client.Join(context.Background(), &proto.JoinRequest{
			JoinToken: body.JoinToken,
		})
		if err != nil {
			log.Printf("[ERR] api: %v", err)
			c.String(http.StatusBadRequest, "Could not perform cluster join operation.")
			return
		}
		if joinRes.Status == proto.Status_UNAUTHORIZED {
			c.String(http.StatusUnauthorized, "That join token is not authorized.")
			return
		}

		// Set up the local properties for ourselves.
		engine.RPCServer.Port = body.RPCPort
		store.RaftPort = body.RaftPort
		store.SerfPort = body.SerfPort
		store.WANSerfPort = body.WANSerfPort

		// Use the address that we got from the server as our local advertise
		// address, and use the generated ID from the remote server as well.
		store.AdvertiseAddr = net.ParseIP(joinRes.AdvertiseAddr)
		store.ID = joinRes.Id

		// Attempt to open the store.
		errCh := make(chan error)
		go func() { errCh <- store.Open() }()
		select {
		case <-store.Started():
		case err := <-errCh:
			log.Printf("[ERR] api: %v", err)
			c.String(http.StatusInternalServerError, "Could not open the store.")
			return
		}

		// The engine is now ready to accept requests, let's save everything we have
		// up until this point and ensure that this config gets maintained. This is
		// because after the store open operation, Raft has started writing it's
		// data to the raft directory, so it's important that we react to this
		// properly.
		engine.Status = StatusReady
		engine.writeConfig()

		// Let the primary server know that we're ready to be joined to it.
		cRes, err := client.ConfirmJoin(context.Background(), &proto.ConfirmJoinRequest{
			RaftAddr: fmt.Sprintf("%s:%d", joinRes.AdvertiseAddr, store.RaftPort),
			Id:       store.ID,
		})
		if err != nil {
			log.Printf("[ERR] api: %v", err)
			c.String(http.StatusBadRequest, "Could not perform cluster join operation.")
			return
		}
		switch cRes.Status {
		case proto.Status_UNAUTHORIZED:
			c.String(http.StatusUnauthorized, "That join token is no longer authorized.")
			return
		case proto.Status_ERROR:
			c.String(http.StatusInternalServerError, "There was an error in joining your node to the store.")
			return
		}

		// Start the RPC server.
		go func() { errCh <- engine.RPCServer.Start() }()
		select {
		case <-engine.RPCServer.Started():
		case err := <-errCh:
			log.Printf("[ERR] store: %s", err)
			c.String(http.StatusInternalServerError, "There was an error starting the RPC server")
			return
		}

		// The join occurred successfully! Update the engine status and config.
		engine.Status = StatusRunning
		engine.writeConfig()

		// Add this node to the list of nodes.
		cmd := command{
			Op:   opNewNode,
			Node: *store.CurrentNode(),
		}
		if err := cmd.Apply(store); err != nil {
			log.Printf("[ERR] store: could not apply new user: %s", err)
			c.String(http.StatusInternalServerError, "Could not add this node to the store state list.")
			return
		}

		c.String(http.StatusOK, "Successfully joined node %s in the cluster.", targetAddr)
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
			Op:   opNewUser,
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

func (s *APIServer) handleListNodes() gin.HandlerFunc {

	type raftState string
	const (
		Worker  raftState = "worker"
		Leader            = "leader"
		Manager           = "manager"
	)

	type node struct {
		// Node identifiers.
		ID      string `json:"id"`
		Address string `json:"address"`

		// Process ports.
		APIPort     int `json:"api_port"` // Not used by Orbit.
		RPCPort     int `json:"rpc_port"`
		RaftPort    int `json:"raft_port"`
		SerfPort    int `json:"serf_port"`
		WANSerfPort int `json:"wan_serf_port"`

		State raftState `json:"state"`
	}

	return func(c *gin.Context) {
		nodes := []node{}
		raftServers := make(map[raft.ServerID]raft.Server)

		// Sort the raft servers by ID.
		configFuture := s.engine.Store.raft.GetConfiguration()
		if err := configFuture.Error(); err != nil {
			c.String(http.StatusInternalServerError, "Could not retrieve raft configuration.")
			return
		}
		for _, srv := range configFuture.Configuration().Servers {
			fmt.Printf("%+v\n", srv)
			raftServers[srv.ID] = srv
		}

		// Construct the list of nodes.
		for _, n := range s.engine.Store.state.Nodes {
			// Compute the state.
			var state raftState
			if _, ok := raftServers[raft.ServerID(n.ID)]; !ok {
				state = Worker // Raft doesn't have this, so Serf must.
			} else if s.engine.Store.raft.Leader() == raft.ServerAddress(fmt.Sprintf("%s:%d", n.Address, n.RaftPort)) {
				state = Leader // This node is the leader of the cluster.
			} else {
				state = Manager // This is node must be a manager of the cluster.
			}

			nodes = append(nodes, node{
				ID:      n.ID,
				Address: n.Address.String(),

				RPCPort:     n.RPCPort,
				RaftPort:    n.RaftPort,
				SerfPort:    n.SerfPort,
				WANSerfPort: n.WANSerfPort,
				State:       state,
			})
		}

		c.JSON(http.StatusOK, &nodes)
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
			Op:   opRemoveUser,
			User: User{ID: id},
		}

		if err := cmd.Apply(store); err != nil {
			log.Printf("[ERR] store: %s", err)
			c.String(http.StatusInternalServerError, "Could not remove that user.")
			return
		}

		c.String(http.StatusOK, "The user has been removed.")
	}
}

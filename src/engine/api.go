package engine

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/hashicorp/raft"
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
	// Register all other routes.
	//

	r.GET("/", s.handleIndex())
	r.GET("/ip", s.handleIP())
	r.GET("/state", s.handleState())
	r.GET("/users", s.handleListUsers())
	r.POST("/setup", s.handleSetup())
	r.POST("/bootstrap", s.handleBootstrap())
	r.POST("/join", s.handleJoin())

	{
		// Routes that require to be the raft leader.
		r := r.Group("")

		r.Use(func(c *gin.Context) {
			if s.engine.Store.raft.State() != raft.Leader {
				c.String(http.StatusInternalServerError, "This node is not the leader of the cluster, and leader forwarding is not yet implemented.")
				c.Abort()
				return
			}

			c.Next()
		})

		r.POST("/signup", s.handleSignup())
		r.DELETE("/user/:id", s.handleRemoveUser())
	}
}

func (s *APIServer) simpleLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Printf("[INFO] api: Received %s at %s", c.Request.Method, c.Request.URL)
		c.Next()
	}
}

func (s *APIServer) handleIndex() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.String(http.StatusOK, "Welcome to the Orbit Engine API.\nAll systems are operational.")
	}
}

func (s *APIServer) handleIP() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip, err := getPublicIP()
		if err != nil {
			c.String(http.StatusInternalServerError, "%s", "Could not retrieve public IP.")
			return
		}
		c.String(http.StatusOK, ip)
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

func (s *APIServer) handleSetup() gin.HandlerFunc {
	engine := s.engine
	store := engine.Store

	type body struct {
		RawIP string `form:"ip" json:"ip"`
	}

	return func(c *gin.Context) {
		var body body
		c.Bind(&body)

		if engine.Status >= Ready {
			c.String(http.StatusBadRequest, "The engine has already been setup.")
			return
		}

		if body.RawIP == "" {
			c.String(http.StatusBadRequest, "You must provide an IP address.")
			return
		}

		ip := net.ParseIP(body.RawIP)
		if ip == nil {
			c.String(http.StatusBadRequest, "The provided IP address is not valid.")
			return
		}

		store.AdvertiseAddr = ip
		engine.writeConfig() // Save the IP address

		// Open the store.
		openErrCh := make(chan error)
		go func() { openErrCh <- store.Open() }()

		// Wait for the store to start or error out.
		select {
		case <-store.Started():
			break
		case err := <-openErrCh:
			c.String(http.StatusInternalServerError, "Could not open the store. Are you sure that the IP address you have provided exists on the node?")
			fmt.Println(err)
			return
		}

		engine.Status = Ready
		engine.writeConfig()
		c.String(http.StatusOK, "The store has been opened successfully.")
	}
}

func (s *APIServer) handleBootstrap() gin.HandlerFunc {
	engine := s.engine
	store := engine.Store

	return func(c *gin.Context) {
		// Ensure that the engine is ready for the bootstrap operation.
		if engine.Status != Ready {
			var msg string
			if engine.Status == Running {
				msg = "The store has already been bootstrapped."
			} else {
				msg = "The store is not ready to be bootstrapped."
			}
			c.String(http.StatusBadRequest, msg)
			return
		}

		// Perform the bootstrap operation.
		if err := store.Bootstrap(); err != nil {
			c.String(http.StatusInternalServerError, "%s.", err)
			return
		}

		// Update the engine status
		engine.Status = Running
		engine.writeConfig() // Save the engine status
		c.String(http.StatusOK, "The server has been successfully bootstrapped.")
	}
}

func (s *APIServer) handleJoin() gin.HandlerFunc {
	engine := s.engine
	store := engine.Store

	type body struct {
		RawAddr string `form:"address" json:"address"` // The raw TCP address of the node.
		NodeID  string `form:"node_id" json:"node_id"` // The ID of the node to join.
	}

	return func(c *gin.Context) {
		var body body
		c.Bind(&body)

		addr, err := net.ResolveTCPAddr("tcp", body.RawAddr)
		if err != nil {
			c.String(http.StatusBadRequest, "The address you have provided is not valid.")
			return
		}

		if err := store.Join(body.NodeID, *addr); err != nil {
			c.String(http.StatusInternalServerError,
				"Could not join the node at '%s' with ID '%s' to this store.",
				body.RawAddr, body.NodeID,
			)
			return
		}

		c.String(http.StatusOK, "Successfully joined that node to the store.")
	}
}

func (s *APIServer) handleSignup() gin.HandlerFunc {
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

		c.String(http.StatusCreated, "New user created.")
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

func (s *APIServer) handleRemoveUser() gin.HandlerFunc {
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

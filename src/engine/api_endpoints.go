package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"orbit.sh/engine/docker"
	"orbit.sh/engine/gluster"

	"github.com/gin-gonic/gin"
	"github.com/hashicorp/raft"
	"google.golang.org/grpc"
	"orbit.sh/engine/proto"
)

func (s *APIServer) handleIndex() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.String(http.StatusOK, "Welcome to the Orbit Engine API.\nAll systems are operational.")
	}
}

func (s *APIServer) handleState() gin.HandlerFunc {
	type res struct {
		Status       Status `json:"status"`        // Engine status int
		StatusString string `json:"status_string"` // Engine status string
		Stage        string `json:"stage"`         // The set up stage
		Mode         string `json:"mode"`          // The set up mode
	}

	return func(c *gin.Context) {
		mode, stage := s.engine.SetupStatus()

		res := res{
			Status:       s.engine.Status,
			StatusString: fmt.Sprintf("%s", s.engine.Status),
			Stage:        stage,
			Mode:         mode,
		}

		c.JSON(http.StatusOK, &res)
	}
}

func (s *APIServer) handleIP() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip, err := getPublicIP()
		if err == nil && ip != nil {
			c.JSON(http.StatusOK, gin.H{"ip": ip.String()})
		} else {
			c.Status(http.StatusNotFound)
		}
	}
}

func (s *APIServer) handleGetTokens() gin.HandlerFunc {
	store := s.engine.Store

	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"manager": store.state.ManagerJoinToken,
			"worker":  store.state.WorkerJoinToken,
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

		res := "Could not boostrap the cluster." // General error response.

		// Ensure that the store can be bootstrapped.
		if engine.Status >= StatusReady {
			c.String(http.StatusConflict, "This node already belongs to a cluster, and cannot be bootstrapped again.")
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
			c.String(http.StatusBadRequest, "Some invalid fields were provided, please check your request and try again.")
			return
		}

		// Validate and parse the provided IP address.
		var advertiseAddr net.IP
		if body.RawAdvertiseAddr != "" {
			ip := net.ParseIP(body.RawAdvertiseAddr)
			if ip == nil {
				c.String(http.StatusBadRequest, "That advertise address is not valid, please make sure it's in the form of an IP address.")
				return
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

			if strings.Contains(err.Error(), "bind: cannot assign requested address") {
				res = "That address could not be used on this node, please ensure that the IP address provided can be used to reach the node."
			}

			c.String(http.StatusInternalServerError, res)
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
			c.String(http.StatusInternalServerError, res)
			return
		}

		// Save the state and set the engine status.
		engine.Status = StatusReady
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
			Node: *store.GenerateNodeDetails(),
		}

		if err := cmd.Apply(store); err != nil {
			log.Printf("[ERR] store: %s", err)
			c.String(http.StatusInternalServerError, "Could not add this node to the list of nodes in the store, despite being joined to it successfully.")
			return
		}

		// Prepare command to create the orbit system namespace.
		namespaceID := store.state.Namespaces.GenerateID()
		cmd = command{
			Op: opNewNamespace,
			Namespace: Namespace{
				ID:   namespaceID,
				Name: "orbit-system",
			},
		}

		if err := cmd.Apply(store); err != nil {
			log.Printf("[ERR] store: %s", err)
			c.String(http.StatusInternalServerError, "Could not add the 'orbit-system' namespace.")
			return
		}

		// Ensure the node is not currently a member of a swarm. If the node is not
		// a member of the swarm, this command will fail. That is completely
		// alright, as it means that we can just carry on anyway.
		docker.ForceLeaveSwarm()

		// Initialise Docker Swarm with the required parameters.
		if err := docker.SwarmInit(advertiseAddr); err != nil {
			c.String(http.StatusInternalServerError, "Could not initialise docker swarm.")
			return
		}

		// Save the join tokens to the store state. This allows them to be used by
		// the future nodes that join.
		cmd = command{
			Op:               opSetJoinTokens,
			ManagerJoinToken: docker.SwarmToken(true),
			WorkerJoinToken:  docker.SwarmToken(false),
		}

		if err := cmd.Apply(store); err != nil {
			log.Printf("[ERR] store: Could not set join tokens in store: %s", err)
			c.String(http.StatusInternalServerError, "Could not set the join tokens on the store.")
			return
		}

		// Create the overlay network for swarm communications.
		if err := docker.CreateOverlayNetwork("orbit"); err != nil {
			c.String(http.StatusInternalServerError, "Could not create orbit overlay network.")
			return
		}

		// Create the volume that Orbit will use for all of its operations.
		vol, err := store.AddVolume(Volume{
			Name:        "repositories-and-registry",
			Size:        1024,
			NamespaceID: namespaceID,
			Bricks: []Brick{Brick{
				NodeID: store.ID,
			}},
		})
		if err != nil {
			log.Printf("[ERR] store: There was an error creating volume %s: %s", vol.ID, err)
			c.String(http.StatusInternalServerError, "Could not create primary orbit data volume.")
			return
		}

		// Publish the registry server with this path in use.
		if err := docker.DeployRegistry(vol.Paths().Data, 6510); err != nil {
			log.Printf("[ERR] docker: There was an error deploying the registry: %s", err)
			c.String(http.StatusInternalServerError, "Could not publish the image registry to Docker Swarm.")
			return
		}

		// Push the console and edge router images to the registry.
		if err := docker.Push("orbit.sh/edge", "orbit.sh/console"); err != nil {
			log.Printf("[ERR] docker: Could not deploy to the registry: %s", err)
			c.String(http.StatusInternalServerError, "Could not push orbit system images to the registry.")
			return
		}

		// Create service declarations for the console and edge router and start
		// them running.
		edgeService := docker.Service{
			Name: "edge",
			Tag:  "orbit.sh/edge",
			Publish: []docker.Publish{
				docker.Publish{Container: 443, Host: 443},
				docker.Publish{Container: 80, Host: 80},
			},
			Mounts: []docker.ServiceMount{
				docker.ServiceMount{
					Source: "/var/run/orbit.sock",
					Target: "/var/run/orbit.sock",
				},
			},
		}
		consoleService := docker.Service{
			Name:    "console",
			Tag:     "orbit.sh/console",
			Publish: []docker.Publish{docker.Publish{Container: 5000, Host: 6500}},
			Mounts: []docker.ServiceMount{
				docker.ServiceMount{
					Source: "/var/run/orbit.sock",
					Target: "/var/run/orbit.sock",
				},
			},
		}
		if err := docker.CreateService(edgeService, consoleService); err != nil {
			log.Printf("[ERR] docker: Could not create service: %s", err)
			c.String(http.StatusInternalServerError, "Could not create the orbit cluster services.")
			return
		}

		c.JSON(http.StatusOK, engine.marshalConfig())
	}
}

func (s *APIServer) handleGetRepositories() gin.HandlerFunc {
	store := s.engine.Store
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, store.state.Repositories)
	}
}

func (s *APIServer) handleRepositoryAdd() gin.HandlerFunc {
	store := s.engine.Store

	type body struct {
		Name      string `form:"name" json:"name"`
		Namespace string `form:"namespace" json:"namespace"`
	}

	return func(c *gin.Context) {
		var body body
		c.ShouldBind(&body)

		// Check the orbit system volume exists.
		volume := store.OrbitSystemVolume()
		if volume == nil {
			c.String(http.StatusServiceUnavailable, "The orbit system volume is not ready for use. Please complete the set up process.")
			return
		}

		// Search for and check the provided repository namespace.
		namespace := store.state.Namespaces.Find(body.Namespace)
		var namespaceID string
		if namespace != nil {
			namespaceID = namespace.ID
		}

		// Create the command for the repository.
		id := store.state.Repositories.GenerateID()
		cmd := command{
			Op: opNewRepository,
			Repository: Repository{
				ID:          id,
				Name:        body.Name,
				NamespaceID: namespaceID,
			},
		}

		// Now apply it to the store.
		if err := cmd.Apply(store); err != nil {
			c.String(http.StatusInternalServerError, "Could not apply the new repository the store.")
			return
		}

		// Actually create the directory that is used.
		path := filepath.Join(volume.Paths().Data, "repositories", id)
		os.MkdirAll(path, 0644)

		c.JSON(http.StatusCreated, cmd.Repository)
	}
}

func (s *APIServer) handleListNamespaces() gin.HandlerFunc {
	store := s.engine.Store

	return func(c *gin.Context) {
		c.JSON(http.StatusOK, store.state.Namespaces)
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

		// Let the primary server know that we're ready to be joined to it.
		cRes, err := client.ConfirmJoin(context.Background(), &proto.ConfirmJoinRequest{
			RaftAddr:  fmt.Sprintf("%s:%d", joinRes.AdvertiseAddr, store.RaftPort),
			Id:        store.ID,
			JoinToken: body.JoinToken,
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

		// Add this node to the list of nodes.
		cmd := command{
			Op:   opNewNode,
			Node: *store.GenerateNodeDetails(),
		}
		if err := cmd.Apply(store); err != nil {
			log.Printf("[ERR] store: Could not apply node: %s", err)
			c.String(http.StatusInternalServerError, "Could not add this node to the store state list.")
			return
		}

		// Ensure the node is not currently a member of a swarm. If the node is not
		// a member of the swarm, this command will fail. That is completely
		// alright, as it means that we can just carry on anyway.
		docker.ForceLeaveSwarm()

		// Join the docker swarm by connecting with the join token.
		if err := docker.JoinSwarm(targetAddr.IP.String(), store.state.ManagerJoinToken); err != nil {
			c.String(http.StatusInternalServerError, "Could not join the docker swarm cluster.")
			return
		}

		engine.Status = StatusReady
		engine.writeConfig()

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

		Profile *multipart.FileHeader `form:"profile" json:"profile"`
	}

	return func(c *gin.Context) {
		var body body
		c.ShouldBind(&body)

		// Read and input the profile file.
		var profile []byte
		if body.Profile != nil {
			file, _ := body.Profile.Open()
			data, _ := ioutil.ReadAll(file)
			profile = data
		}

		// Construct the user.
		newUser, err := store.state.Users.Generate(UserConfig{
			Name:     body.Name,
			Password: body.Password,
			Username: body.Username,
			Email:    body.Email,
			Profile:  profile,
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
			log.Printf("[ERR] store: Could not perform apply: %s", err)
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

// Deliver the user profile as an image.
func (s *APIServer) handleUserProfile() gin.HandlerFunc {
	store := s.engine.Store

	return func(c *gin.Context) {
		id := c.Param("id")

		// Search for the user with that ID, username, or email address.
		var user *User
		for _, u := range store.state.Users {
			if u.ID == id || u.Email == id || u.Username == id {
				user = &u
			}
		}
		if user == nil {
			c.String(http.StatusNotFound, "A user with the identifier '%s' could not be found.", id)
			return
		}

		// If there is no profile data, ensure we return a 404.
		if len(user.Profile) == 0 {
			c.String(http.StatusNoContent, "The user was found but they have no profile image.")
			return
		}

		// Send the profile image data. This will also take in the MIME type of the
		// byte slice and automatically decode it to the correct one.
		profile := user.Profile
		ct := http.DetectContentType(profile)
		c.Data(http.StatusOK, ct, profile)
	}
}

func (s *APIServer) handleListNodes() gin.HandlerFunc {
	engine := s.engine
	store := engine.Store

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
		RPCPort     int `json:"rpc_port"`
		RaftPort    int `json:"raft_port"`
		SerfPort    int `json:"serf_port"`
		WANSerfPort int `json:"wan_serf_port"`

		// Add the node specific details.
		Roles      []NodeRole `json:"node_roles"`
		SwapSize   int        `json:"swap_size"`
		Swappiness int        `json:"swappiness"`

		// Add the raft state
		State raftState `json:"state"`
	}

	return func(c *gin.Context) {
		if engine.Status < StatusReady {
			c.JSON(http.StatusBadRequest, []interface{}{})
			return
		}

		nodes := []node{}
		raftServers := make(map[raft.ServerID]raft.Server)

		// Sort the raft servers by ID.
		configFuture := store.raft.GetConfiguration()
		if err := configFuture.Error(); err != nil {
			c.String(http.StatusInternalServerError, "Could not retrieve raft configuration.")
			return
		}
		for _, srv := range configFuture.Configuration().Servers {
			raftServers[srv.ID] = srv
		}

		// Construct the list of nodes.
		for _, n := range store.state.Nodes {
			// Compute the state.
			var state raftState
			if _, ok := raftServers[raft.ServerID(n.ID)]; !ok {
				state = Worker // Raft doesn't have this, so Serf must.
			} else if store.raft.Leader() == raft.ServerAddress(fmt.Sprintf("%s:%d", n.Address, n.RaftPort)) {
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

				Roles:      n.Roles,
				SwapSize:   n.SwapSize,
				Swappiness: n.Swappiness,

				State: state,
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

func (s *APIServer) handleSnapshot() gin.HandlerFunc {
	return func(c *gin.Context) {
		op := c.Param("op")
		switch op {
		case "take":
			// Take a snapshot.
			f := s.engine.Store.raft.Snapshot()
			if err := f.Error(); err != nil {
				c.String(http.StatusInternalServerError, "Error processing the snapshot: %s", err)
				return
			}

			// Attempt to open the snapshot.
			_, rc, err := f.Open()
			if err != nil {
				c.String(http.StatusInternalServerError, "Error opening the snapshot: %s", err)
				return
			}

			// Plop the snapshot in JSON and send to the user.
			snapshot := &fsmSnapshot{}
			if err := json.NewDecoder(rc).Decode(snapshot); err != nil {
				c.String(http.StatusInternalServerError, "Error decoding the snapshot: %s", err)
				return
			}
			c.JSON(http.StatusOK, snapshot)

		case "restore":
			c.String(
				http.StatusNotImplemented,
				`The restore method of the snapshot hasn't been implemented.
This is because it is for testing purposes.`,
			)

		default:
			c.String(http.StatusBadRequest, "Operation must be 'take' or 'restore'.")
			return
		}
	}
}

func (s *APIServer) handleListRouters() gin.HandlerFunc {
	store := s.engine.Store

	return func(c *gin.Context) {
		store.mu.RLock()
		defer store.mu.RUnlock()

		c.JSON(http.StatusOK, store.state.Routers)
	}
}

// This will add a router object to the store.
func (s *APIServer) handleRouterAdd() gin.HandlerFunc {
	store := s.engine.Store

	type body struct {
		Domain      string `form:"domain" json:"domain"`
		Namespace   string `form:"namespace" json:"namespace"`
		AppID       string `form:"app_id" json:"app_id"`
		WWWRedirect bool   `form:"www_redirect" json:"www_redirect"`
	}

	return func(c *gin.Context) {
		var body body
		c.Bind(&body)

		// Generate the ID for the router.
		id := store.state.Routers.GenerateID()

		// Find the namespace by ID.
		var namespaceID string
		namespace := store.state.Namespaces.Find(body.Namespace)
		if namespace != nil {
			namespaceID = namespace.ID
		}

		// Create a new router without a certificate.
		cmd := command{
			Op: opNewRouter,
			Router: Router{
				ID:          id,
				Domain:      body.Domain,
				NamespaceID: namespaceID,
				AppID:       body.AppID,
				WWWRedirect: body.WWWRedirect,
			},
		}

		// Actually create the router.
		if err := cmd.Apply(store); err != nil {
			log.Printf("[ERR] store: %s", err)
			c.String(http.StatusInternalServerError, "Could not create that router.")
			return
		}

		c.String(http.StatusCreated, id)
	}
}

// Update a router object.
func (s *APIServer) handleRouterUpdate() gin.HandlerFunc {
	store := s.engine.Store

	type body struct {
		CertificateID string `form:"certificate_id" json:"certificate_id"`
		Namespace     string `form:"namespace" json:"namespace"`
		AppID         string `form:"app_id" json:"app_id"`
	}

	return func(c *gin.Context) {
		id := c.Param("id") // The ID of the router to update
		var body body
		c.Bind(&body)

		// Find the namespace by ID. The namespace doesn't need to exist, so if this
		// doesn't work, just don't set the namespace to be updated.
		namespace := store.state.Namespaces.Find(body.Namespace)
		namespaceID := ""
		if namespace != nil {
			namespaceID = namespace.ID
		}

		// Create the update command.
		cmd := command{
			Op: opUpdateRouter,
			Router: Router{
				ID:            id,
				CertificateID: body.CertificateID,
				NamespaceID:   namespaceID,
				AppID:         body.AppID,
			},
		}

		// Attempt to apply the update command.
		if err := cmd.Apply(store); err != nil {
			log.Printf("[ERR] store: %s", err)
			c.String(http.StatusInternalServerError, "Could not update the router.")
			return
		}

		// This was successful.
		c.String(http.StatusOK, "Successfully updated your router.")
	}
}

func (s *APIServer) handleListCertificates() gin.HandlerFunc {
	store := s.engine.Store

	return func(c *gin.Context) {
		store.mu.RLock()
		defer store.mu.RUnlock()

		c.JSON(http.StatusOK, store.state.Certificates)
	}
}

// This will add a certificate to the store. Either raw certificate data can be
// uploaded, or it can be enabled for auto renewal so that certificate data
// doesn't need to be uploaded.
func (s *APIServer) handleCertificateAdd() gin.HandlerFunc {
	store := s.engine.Store

	type body struct {
		AutoRenew bool     `form:"auto_renew" json:"auto_renew"`
		Namespace string   `form:"namespace" json:"namespace"`
		Domains   []string `form:"domains" json:"domains"`

		PrivateKey *multipart.FileHeader `form:"private_key" json:"private_key"`
		FullChain  *multipart.FileHeader `form:"full_chain" json:"full_chain"`
	}

	return func(c *gin.Context) {
		var body body
		c.ShouldBind(&body)

		// Read and parse the full chain and private key.
		var fullChain, privateKey []byte
		if body.FullChain != nil {
			file, _ := body.FullChain.Open()
			data, _ := ioutil.ReadAll(file)
			fullChain = data
		}
		if body.PrivateKey != nil {
			file, _ := body.PrivateKey.Open()
			data, _ := ioutil.ReadAll(file)
			privateKey = data
		}

		// Generate the certificate ID.
		id := store.state.Certificates.GenerateID()

		// Find the namespace by ID.
		namespace := store.state.Namespaces.Find(body.Namespace)
		if namespace == nil {
			c.String(http.StatusNotFound, "No namespace with the name or ID %s could be found.", body.Namespace)
			return
		}

		// Construct the command.
		cmd := command{
			Op: opNewCertificate,
			Certificate: Certificate{
				ID:          id,
				AutoRenew:   body.AutoRenew,
				FullChain:   fullChain,
				PrivateKey:  privateKey,
				NamespaceID: namespace.ID,
				Domains:     body.Domains,
			},
		}

		// Ensure that the data is correct.
		if !cmd.Certificate.AutoRenew && (len(cmd.Certificate.FullChain) == 0 || len(cmd.Certificate.PrivateKey) == 0) {
			log.Printf("[INFO] api: Neither auto renew or certificate supplied")
			c.String(http.StatusBadRequest, "You must supply either auto renew or certificate data.")
			return
		}

		// Apply the certificate to the store.
		if err := cmd.Apply(store); err != nil {
			log.Printf("[ERR]: store: %s", err)
			c.String(http.StatusInternalServerError, "Could not add the certificate to the store.")
			return
		}

		// Otherwise, return the ID of the generated certificate.
		c.String(http.StatusCreated, id)
	}
}

func (s *APIServer) handleRenewCertificates() gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := s.engine.Store.RenewCertificates(); err != nil {
			log.Printf("[ERR] api: Could not renew certificates: %s", err)
			c.String(http.StatusInternalServerError, "Could not renew certificates")
			return
		}
	}
}

func (s *APIServer) handleRestartService() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if err := docker.ForceUpdateService(id); err != nil {
			log.Printf("[ERR] api: Could not force update service: %s", err)
			c.String(http.StatusInternalServerError, "Could not force update the %s service.", id)
			return
		}
		c.String(http.StatusOK, "Force updated the %s service.", id)
	}
}

func (s *APIServer) handleGetNode() gin.HandlerFunc {
	store := s.engine.Store

	return func(c *gin.Context) {
		// Retrieve the ID, and if the ID is "current", that acts a shorthand that
		// refers to the node that the API server is running and receiving requests
		// for (this instance, essentially).
		id := c.Param("id")
		if id == "current" {
			id = store.ID
		}

		// Retrieve the node with this ID.
		var node *Node
		for _, n := range store.state.Nodes {
			if n.ID == id {
				node = &n
				break
			}
		}
		if node == nil {
			c.String(http.StatusNotFound, "Could not find a node with that ID.")
			return
		}

		// Return the node details.
		c.JSON(http.StatusOK, node)
	}
}

func (s *APIServer) handleUserGet() gin.HandlerFunc {
	store := s.engine.Store

	return func(c *gin.Context) {
		// This identifier will search for a user who has this kind of identifying
		// information. This can be a username, email address, ID, or session token.
		id := c.Param("id")

		// Find the user with that identifier.
		var user *User
	search:
		for _, u := range store.state.Users {
			// Search by that user's flat properties.
			if u.ID == id || u.Username == id || u.Email == id {
				user = &u
				break search
			}

			// Search their session tokens.
			for _, s := range u.Sessions {
				if s.Token == id {
					user = &u
					break search
				}
			}
		}
		if user == nil {
			c.String(http.StatusNotFound, "A user with those details could not be found.")
			return
		}

		// Sanitize and send the user object.
		c.JSON(http.StatusOK, gin.H{
			"id":       user.ID,
			"username": user.Username,
			"email":    user.Email,
			"name":     user.Name,
		})
	}
}

func (s *APIServer) handleUserLogin() gin.HandlerFunc {
	store := s.engine.Store

	type body struct {
		Identifier string `form:"identifier" json:"identifier"`
		Password   string `form:"password" json:"password"`
	}

	return func(c *gin.Context) {
		// Bind the identifier and username details.
		var body body
		if err := c.Bind(&body); err != nil {
			c.String(http.StatusBadRequest, "You must supply an identifier and a password.")
			return
		}

		// Search for the credentials that match this user.
		var user *User
		for _, u := range store.state.Users {
			if u.Username == body.Identifier || u.Email == body.Identifier {
				user = &u
				break
			}
		}
		if user == nil {
			c.String(http.StatusNotFound, "That user doesn't exist.")
			return
		}

		// Check if the user credentials match.
		if !user.ValidatePassword(body.Password) {
			c.String(http.StatusUnauthorized, "The password you provided is incorrect.")
			return
		}

		// Otherwise, we can now log the user in by generating a session token.
		cmd := command{
			Op:      opNewSession,
			User:    User{ID: user.ID},
			Session: user.GenerateSession(),
		}

		// Apply the session to the store.
		if err := cmd.Apply(store); err != nil {
			log.Fatalf("[ERR] store: Could not apply new session to user: %s", err)
			c.String(http.StatusInternalServerError, "Can't update store.")
			return
		}

		// Wait for the application to take place.
	search:
		for {
			for _, u := range store.state.Users {
				for _, s := range u.Sessions {
					if s.Token == cmd.Session.Token {
						break search
					}
				}
			}
			time.Sleep(time.Millisecond * 200)
		}

		// Return the session token to the user.
		c.String(http.StatusOK, cmd.Session.Token)
	}
}

func (s *APIServer) handleSessionRevoke() gin.HandlerFunc {
	store := s.engine.Store

	return func(c *gin.Context) {
		id := c.Param("id")
		token := c.Param("token")

		// Find the user for this request.
		var user *User
		for _, u := range store.state.Users {
			if u.Username == id || u.Email == id || u.ID == id {
				user = &u
				break
			}
		}
		if user == nil {
			c.String(http.StatusNotFound, "That user doesn't exist.")
			return
		}

		// Revoke all sessions.
		if token == "all" {
			cmd := command{
				Op:   opRevokeAllSessions,
				User: User{ID: user.ID},
			}

			if err := cmd.Apply(store); err != nil {
				log.Printf("[ERR] store: Could not revoke all sessions: %s", err)
				c.String(http.StatusInternalServerError, "Could not revoke all sessions.")
				return
			}
		} else {
			// Revoke that individual session.
			cmd := command{
				Op:      opRevokeSession,
				Session: Session{Token: token},
			}

			if err := cmd.Apply(store); err != nil {
				log.Printf("[ERR] store: Could not revoke session: %s", err)
				c.String(http.StatusInternalServerError, "Could not revoke that session.")
				return
			}
		}

		c.String(http.StatusOK, "Session(s) revoked.")
	}
}

func (s *APIServer) handleNodeUpdate() gin.HandlerFunc {
	engine := s.engine
	store := engine.Store

	type body struct {
		NodeRoles  []NodeRole `form:"node_roles" json:"node_roles"`
		SwapSize   int        `form:"swap_size" json:"swap_size"`
		Swappiness int        `form:"swappiness" json:"swappiness"`
	}

	return func(c *gin.Context) {
		// Retrieve the ID, and if the ID is "current", that acts a shorthand that
		// refers to the node that the API server is running and receiving requests
		// for (this instance, essentially).
		id := c.Param("id")
		if id == "current" {
			id = store.ID
		}

		// Retrieve the body values (the actual update values).
		var body body
		if err := c.ShouldBind(&body); err != nil {
			log.Printf("[ERR] api: Could not bind node update values: %s", err)
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		// Construct the store update expression.
		cmd := command{
			Op: opUpdateNode,
			Node: Node{
				ID:         id,
				SwapSize:   body.SwapSize,
				Swappiness: body.Swappiness,
				Roles:      body.NodeRoles,
			},
		}

		if err := cmd.Apply(store); err != nil {
			log.Printf("[ERR] api: Could not update store: %s", err)
			c.String(http.StatusInternalServerError, "Could not update the store.")
			return
		}

		// As this route can also be used as the last stage of the set up process,
		// let's check if the node included a manager or worker role update status.
		// If it did, we can successfully classify the engine as running (assuming
		// that it isn't already).
		if engine.Status != StatusRunning &&
			(cmd.Node.HasRole(RoleManager) || cmd.Node.HasRole(RoleWorker)) {
			engine.Status = StatusRunning
			engine.writeConfig() // Update the config file with new status
		}

		c.String(http.StatusOK, "Successfully updated the node with id %s", id)
	}
}

func (s *APIServer) handleVolumeAdd() gin.HandlerFunc {
	store := s.engine.Store

	type body struct {
		Name   string   `form:"name" json:"name"`
		Size   int      `form:"size" json:"size"`
		Bricks []string `form:"bricks" json:"bricks"`

		Namespace string `form:"namespace" json:"namespace"`
	}

	return func(c *gin.Context) {
		var body body
		if err := c.Bind(&body); err != nil {
			c.String(http.StatusBadRequest, "Invalid body fields.")
			return
		}

		// Convert the body bricks into actual bricks.
		var bricks []Brick
		for _, b := range body.Bricks {
			// Find the node that represents this brick name.
			for _, n := range store.state.Nodes {
				if n.ID == b || n.Address.String() == b {
					// The node was found, create the brick.
					bricks = append(bricks, Brick{
						NodeID:  n.ID,
						Created: false,
					})

					// This brick has found its node, we can break this loop and continue
					// with the rest of the bricks.
					break
				}
			}
		}

		// Sanity check that the desired bricks completely match the server outcome.
		// If there is a disparity between these two values, then it means that
		// there one of the brick node names provided was invalid.
		if len(bricks) != len(body.Bricks) {
			log.Printf("[ERR] api: Invalid brick request. Found %d nodes out of the provided %d bricks.", len(bricks), len(body.Bricks))
			c.String(http.StatusBadRequest, "One of the bricks you provided doesn't exist.")
			return
		}

		// Find the namespace by ID.
		namespace := store.state.Namespaces.Find(body.Namespace)
		if namespace == nil {
			c.String(http.StatusNotFound, "No namespace with the name or ID %s could be found.", body.Namespace)
			return
		}

		// Construct the volume.
		volume := Volume{
			Name:        body.Name,
			Size:        body.Size,
			Bricks:      bricks,
			NamespaceID: namespace.ID,
		}

		// Add the volume to the store.
		v, err := store.AddVolume(volume)
		if err != nil {
			log.Printf("[ERR] api: Could not create the volume: %s", err)
			c.String(http.StatusBadRequest, "That volume could not be created.")
		}

		c.JSON(http.StatusCreated, v)
	}
}

func (s *APIServer) handleVolumeRemove() gin.HandlerFunc {
	store := s.engine.Store
	return func(c *gin.Context) {
		id := c.Param("id")

		// Find the volume in question by name and ID.
		var volume *Volume
		for _, v := range store.state.Volumes {
			if id == v.ID || id == v.Name {
				volume = &v
				break
			}
		}

		// If there was no volume, notify the user.
		if volume == nil {
			c.String(http.StatusNotFound, "A volume with the name or ID '%s' does not exist.", id)
			return
		}

		// Otherwise, create and apply the remove operation.
		cmd := command{
			Op:     opRemoveVolume,
			Volume: Volume{ID: volume.ID},
		}

		if err := cmd.Apply(store); err != nil {
			log.Printf("[ERR] store: Could not apply the volume remove operation: %s", err)
			c.String(http.StatusInternalServerError, "Could not apply store remove operation.")
			return
		}

		// Stop and delete the volume with that ID.
		gluster.StopVolume(id)
		gluster.DeleteVolume(id)

		// Return the removed ID along with confirmation.
		c.String(http.StatusOK, volume.ID)
	}
}

func (s *APIServer) handleListVolumes() gin.HandlerFunc {
	store := s.engine.Store

	return func(c *gin.Context) {
		c.JSON(http.StatusOK, store.state.Volumes)
	}
}

func (s *APIServer) handleListDeployments() gin.HandlerFunc {
	store := s.engine.Store
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, store.state.Deployments)
	}
}

func (s *APIServer) handleDeploymentAdd() gin.HandlerFunc {
	engine := s.engine
	store := engine.Store

	type body struct {
		Name         string `form:"name" json:"name"`
		RepositoryID string `form:"repository_id" json:"repository_id"`
		Path         string `form:"path" json:"path"`
		Branch       string `form:"branch" json:"branch"`
		Namespace    string `form:"namespace" json:"namespace"`
	}

	return func(c *gin.Context) {
		var body body
		c.ShouldBind(&body)

		// Ensure that there is a name and a repo.
		if body.RepositoryID == "" || body.Name == "" {
			c.String(http.StatusBadRequest, "Need to provide a repository_id and name.")
			return
		}

		// Search for the namespace.
		var namespaceID string
		namespace := store.state.Namespaces.Find(body.Namespace)
		if namespace != nil {
			namespaceID = namespace.ID
		}

		// Construct the create command and apply it.
		id := store.state.Deployments.GenerateID()
		cmd := command{
			Op: opNewDeployment,
			Deployment: Deployment{
				ID:          id,
				Name:        body.Name,
				Repository:  body.RepositoryID,
				Path:        body.Path,
				NamespaceID: namespaceID,
			},
		}

		if err := cmd.Apply(store); err != nil {
			log.Printf("[ERR] store: Could not apply the deployment to the store: %s", err)
			c.String(http.StatusInternalServerError, "Could not apply the deployment to the store.")
			return
		}

		// Otherwise on success just return the created deployment ID.
		c.String(http.StatusCreated, id)
	}
}

func (s *APIServer) handleBuildDeployment() gin.HandlerFunc {
	engine := s.engine
	store := engine.Store

	return func(c *gin.Context) {
		id := c.Param("id")

		// Find the ID of the deployment provided.
		var deployment *Deployment
		for _, d := range store.state.Deployments {
			if d.ID == id {
				deployment = &d
				break
			}
		}
		if deployment == nil {
			c.String(http.StatusNotFound, "No deployment with that ID exists.")
			return
		}

		// Now run the build process.
		key, err := engine.BuildDeployment(*deployment)
		if err != nil {
			log.Printf("[ERR] deployment: %s", err)
			c.String(http.StatusInternalServerError, "Could not build deployment.")
			return
		}

		// Create a shorthand function log to the build log entries for this
		// deployment.
		buildLog := func(format string, values ...interface{}) {
			str := fmt.Sprintf(format, values...)
			store.AppendBuildLog(deployment.ID, key, str)
		}

		// Push the image to the docker image registry.
		buildLog("Pushing image %s to the local docker registry", deployment.ID)
		if err := docker.Push(deployment.ID); err != nil {
			log.Printf("[ERR] deployment: %s", err)
			c.String(http.StatusInternalServerError, "Could not push to docker image registry.")
			return
		}
		buildLog("Image %s pushed successfully", deployment.ID)

		// Ensure that the service gets removed correctly.
		if existed := docker.RemoveService(deployment.ID); existed {
			buildLog("Removed existing service %s", deployment.ID)
		}

		// Create the service.
		buildLog("Creating the docker service definition for %s", deployment.ID)
		service := docker.Service{
			Name:    deployment.ID,
			Tag:     deployment.ID,
			Command: "/start",
			Args:    []string{"web"},
		}
		if err := docker.CreateService(service); err != nil {
			log.Printf("[ERR] deployment: %s", err)
			c.String(http.StatusInternalServerError, "Could not create docker service.")
			return
		}
		buildLog("Docker service %s created", deployment.ID)

		// The deployment process has finished.
		buildLog("-----> Deployment succeeded!")
		c.String(http.StatusCreated, deployment.ID)
	}
}

func (s *APIServer) handleRouterRemove() gin.HandlerFunc {
	store := s.engine.Store

	return func(c *gin.Context) {
		id := c.Param("id")

		// Find the ID provided.
		var router *Router
		for _, r := range store.state.Routers {
			if r.ID == id {
				router = &r
				break
			}
		}
		if router == nil {
			c.String(http.StatusNotFound, "Router with the ID of %s could not be found.", id)
			return
		}

		// Perform the delete apply operation.
		cmd := command{
			Op:     opRemoveRouter,
			Router: Router{ID: id},
		}
		if err := cmd.Apply(store); err != nil {
			log.Printf("[ERR] store: %s", err)
			c.String(http.StatusInternalServerError, "Could not apply router removal to the store.")
			return
		}

		c.String(http.StatusOK, id)
	}
}

func (s *APIServer) handleCertificateRemove() gin.HandlerFunc {
	store := s.engine.Store

	return func(c *gin.Context) {
		id := c.Param("id")

		// Find the ID provided.
		var certificate *Certificate
		for _, c := range store.state.Certificates {
			if c.ID == id {
				certificate = &c
				break
			}
		}
		if certificate == nil {
			c.String(http.StatusNotFound, "Certificate with the ID of %s could not be found.", id)
			return
		}

		// Perform the delete apply operation.
		cmd := command{
			Op:          opRemoveCertificate,
			Certificate: Certificate{ID: id},
		}
		if err := cmd.Apply(store); err != nil {
			log.Printf("[ERR] store: %s", err)
			c.String(http.StatusInternalServerError, "Could not apply certificate removal to the store.")
			return
		}

		c.String(http.StatusOK, id)
	}
}

func (s *APIServer) handleNodeRemove() gin.HandlerFunc {
	return func(c *gin.Context) {

	}
}

func (s *APIServer) handleRepositoryRemove() gin.HandlerFunc {
	return func(c *gin.Context) {

	}
}

func (s *APIServer) handleDeploymentRemove() gin.HandlerFunc {
	return func(c *gin.Context) {

	}
}

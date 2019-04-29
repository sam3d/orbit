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
	"strings"
	"sync"
	"time"

	"orbit.sh/engine/docker"

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
			Node: *store.CurrentNode(),
		}

		if err := cmd.Apply(store); err != nil {
			log.Printf("[ERR] store: %s", err)
			c.String(http.StatusInternalServerError, "Could not add this node to the list of nodes in the store, despite being joined to it successfully.")
			return
		}

		// Prepare command to create the orbit system namespace.
		cmd = command{
			Op: opNewNamespace,
			Namespace: Namespace{
				ID:   store.state.Namespaces.GenerateID(),
				Name: "orbit-system",
			},
		}

		if err := cmd.Apply(store); err != nil {
			log.Printf("[ERR] store: %s", err)
			c.String(http.StatusInternalServerError, "Could not add the 'orbit-system' namespace.")
			return
		}

		c.JSON(http.StatusOK, engine.marshalConfig())
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
		NamespaceID string `form:"namespace_id" json:"namespace_id"`
		AppID       string `form:"app_id" json:"app_id"`
	}

	return func(c *gin.Context) {
		var body body
		c.Bind(&body)

		// Generate the ID for the router.
		id := store.state.Routers.GenerateID()

		// Create a new router without a certificate.
		cmd := command{
			Op: opNewRouter,
			Router: Router{
				ID:          id,
				Domain:      body.Domain,
				NamespaceID: body.NamespaceID,
				AppID:       body.AppID,
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
		NamespaceID   string `form:"namespace_id" json:"namespace_id"`
		AppID         string `form:"app_id" json:"app_id"`
	}

	return func(c *gin.Context) {
		id := c.Param("id") // The ID of the router to update
		var body body
		c.Bind(&body)

		// Create the update command.
		cmd := command{
			Op: opUpdateRouter,
			Router: Router{
				ID:            id,
				CertificateID: body.CertificateID,
				NamespaceID:   body.NamespaceID,
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
		AutoRenew   bool                  `form:"auto_renew" json:"auto_renew"`
		FullChain   *multipart.FileHeader `form:"full_chain" json:"full_chain"`
		PrivateKey  *multipart.FileHeader `form:"private_key" json:"private_key"`
		NamespaceID string                `form:"namespace_id" json:"namespace_id"`
		Domains     []string              `form:"domains" json:"domains"`
	}

	return func(c *gin.Context) {
		var body body
		c.Bind(&body)

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

		// Construct the command.
		cmd := command{
			Op: opNewCertificate,
			Certificate: Certificate{
				ID:          id,
				AutoRenew:   body.AutoRenew,
				FullChain:   fullChain,
				PrivateKey:  privateKey,
				NamespaceID: body.NamespaceID,
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

func (s *APIServer) handleRestartService() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if err := docker.ForceUpdateService(id); err != nil {
			log.Printf("[ERR] api: Could not force update service: %s", err)
			c.String(http.StatusInternalServerError, "Could not force update the %s service.", id)
			return
		}
		c.String(http.StatusOK, "Force updated the %s service", id)
	}
}

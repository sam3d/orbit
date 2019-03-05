package engine

import (
	"fmt"
	"log"
	"net"
	"net/http"

	"github.com/gin-gonic/gin"
)

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

	r.POST("/setup", s.handleSetup())
	r.POST("/bootstrap", s.handleBootstrap())
	r.POST("/join", s.handleJoin())
}

func (s *APIServer) simpleLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		go func() {
			log.Printf("[INFO] api: Received %s at %s", c.Request.Method, c.Request.URL)
		}()

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

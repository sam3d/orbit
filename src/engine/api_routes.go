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

	// Register middleware
	r.Use(s.simpleLogger())

	// Register custom routes
	r.GET("/", s.handleIndex())
	r.GET("/state", s.handleState())
	r.POST("/bootstrap", s.handleBootstrap())
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

func (s *APIServer) handleState() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":        s.engine.Status,
			"status_string": fmt.Sprintf("%s", s.engine.Status),
		})
	}
}

func (s *APIServer) handleBootstrap() gin.HandlerFunc {
	engine := s.engine
	store := engine.Store

	type body struct {
		RawIP string `json:"ip"`
	}

	return func(c *gin.Context) {
		var body body
		c.BindJSON(&body)

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
			c.String(http.StatusInternalServerError, "Could not open the store.")
			fmt.Println(err)
			return
		}

		if err := store.Bootstrap(); err != nil {
			c.String(http.StatusInternalServerError, "%s.", err)
			return
		}

		// Update the engine status
		engine.Status = Ready
		engine.writeConfig() // Save the engine status
		c.String(http.StatusOK, "The server has been successfully bootstrapped.")
	}
}

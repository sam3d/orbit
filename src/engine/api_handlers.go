package engine

import (
	"log"
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
	// Handle all of the routes.
	//

	r.GET("", s.handleIndex())

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

func (s *APIServer) handleIndex() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.String(http.StatusOK, "Welcome to the Orbit Engine API.\nAll systems are operational.")
	}
}

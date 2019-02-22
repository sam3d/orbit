package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// routes registers all of the default routes for the API server. This is a
// separate method so that other routes can be added *after* the defaults but
// *before* the server is started.
func (s *Server) routes() {
	r := s.router

	r.GET("/", s.handleIndex())
	r.GET("/state", s.handleState())
}

func (s *Server) handleIndex() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.String(http.StatusOK, "Welcome to the Orbit Engine API.\nAll systems are operational.")
	}
}

func (s *Server) handleState() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	}
}

package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// routes registers all of the default routes for the API server. This is a
// separate method so that other routes can be added *after* the defaults but
// *before* the server is started.
func (s *Server) routes() {
	s.router.Handle("GET", "/", s.handleIndex())
	s.router.Handle("GET", "/state", s.handleState())
}

func (s *Server) handleIndex() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.String(http.StatusOK, "Hello, world!")
	}
}

func (s *Server) handleState() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	}
}

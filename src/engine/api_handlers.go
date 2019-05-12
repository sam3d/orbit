package engine

import (
	"log"

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
	r.GET("/ip", s.handleIP())

	// Group list gets.
	r.GET("/users", s.handleListUsers())
	r.GET("/nodes", s.handleListNodes())
	r.GET("/routers", s.handleListRouters())
	r.GET("/certificates", s.handleListCertificates())
	r.GET("/namespaces", s.handleListNamespaces())

	r.POST("/snapshot/:op", s.handleSnapshot())
	r.POST("/service/restart/:id", s.handleRestartService())
	r.POST("/certificates/renew", s.handleRenewCertificates())

	{
		r := r.Group("/cluster")
		r.POST("/bootstrap", s.handleClusterBootstrap())
		r.POST("/join", s.handleClusterJoin())
	}

	{
		r := r.Group("/user")
		r.POST("", s.handleUserSignup())
		r.GET("/:id/profile", s.handleUserProfile())
		r.GET("/:id", s.handleUserGet())
		r.DELETE("/:id", s.handleUserRemove())
	}

	{
		r := r.Group("/router")
		r.POST("", s.handleRouterAdd())
		r.PUT("/:id", s.handleRouterUpdate())
	}

	{
		r := r.Group("/certificate")
		r.POST("", s.handleCertificateAdd())
	}

	{
		r := r.Group("/node")
		r.GET("/:id", s.handleGetNode())
		r.PUT("/:id", s.handleNodeUpdate())
	}
}

func (s *APIServer) simpleLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Printf("[INFO] api: Received %s at %s", c.Request.Method, c.Request.URL)
		c.Next()
	}
}

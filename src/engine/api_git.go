package engine

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sosedoff/gitkit"
)

// handleGit will return a handler for HTTP git requests on the HTTP server.
func (s *APIServer) handleGit() gin.HandlerFunc {
	store := s.engine.Store
	var service *gitkit.Server

	// Create a separate handler that this launches. This is required so that we
	// can properly ensure that the store is set up before we are able to perform
	// any of the git functions.
	return func(c *gin.Context) {
		notReady := func() {
			c.String(http.StatusServiceUnavailable, "Orbit is not yet ready to handle requests.\nPlease complete the set up process.")
		}

		// Ensure that the engine is ready to receive git requests (essentially, if
		// there's a volume that uses the orbit-system namespace).
		namespace := store.state.Namespaces.Find("orbit-system")
		if namespace == nil {
			notReady()
			return
		}
		var volume *Volume
		for _, v := range store.state.Volumes {
			if v.NamespaceID == namespace.ID {
				volume = &v
				break
			}
		}
		if volume == nil {
			notReady()
			return
		}

		// The system is ready, if the service has not yet been instantiated, make
		// sure that it gets instantiated.
		if service == nil {
			service = gitkit.New(gitkit.Config{Auth: true})

			// The following function is responsible for performing all of the checks
			// on the URL and ensuring that it ends up in the place that it's expected
			// to.
			service.AuthFunc = func(creds gitkit.Credential, req *gitkit.Request) (bool, error) {
				// Find the user attached.
				user := store.state.Users.Find(creds.Username)
				if user == nil {
					// That user does not exist.
					return false, nil
				}
				if !user.ValidatePassword(creds.Password) {
					// The user's password is incorrect.
					return false, nil
				}

				// Find the repo based on the given URL.
				urlPath := strings.TrimPrefix(req.RepoPath, "repo/")
				fmt.Println(urlPath)

				// Check if it exists against the store.state.Repositories
				// Set the req.RepoPath to be the desired location in the correct volume.
				req.RepoPath = urlPath

				return true, nil
			}
		}

		// Now continue executing the service as though it was gin middleware.
		service.ServeHTTP(c.Writer, c.Request)
	}
}

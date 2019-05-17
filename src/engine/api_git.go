package engine

import (
	"log"
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

				// Remove the repo prefix from the URL so that it's not a factor.
				urlPath := strings.TrimPrefix(req.RepoPath, "repo/")

				// Attempt to split the path into two items. If there's two items, it
				// means that the repo is referenced by its name and namespace, and if
				// there's only one item, it means that it's referenced by its name in
				// the "default" namespace or its unique identifier in any namespace.
				tokens := strings.Split(urlPath, "/")

				var namespace *Namespace
				var identifier string
				switch len(tokens) {
				case 1:
					identifier = tokens[0]
				case 2:
					namespace = store.state.Namespaces.Find(tokens[0])
					identifier = tokens[1]
				default:
					log.Printf("[ERR] git: Wrong number of URL components")
					return false, nil
				}

				// Search through the repositories to find matching ones.
				var repo *Repository
				for _, r := range store.state.Repositories {
					// Check the ID first.
					if r.ID == identifier {
						repo = &r
						break
					}

					// If there was no namespace provided, just return the first match for
					// the given repository name. This handles the case of the "default"
					// repository. If no match is provided, continue on with the loop
					// anyway, as the following checks require there to be a namespace.
					if namespace == nil {
						if r.Name == identifier {
							repo = &r
							break
						}
						continue
					}

					// And finally, if the namespace and name match, then we can return
					// the repo for that result.
					if r.NamespaceID == namespace.ID && r.Name == identifier {
						repo = &r
						break
					}
				}

				// If there was no repository found by this point, it means that with
				// the details provided, there wasn't a single one found.
				if repo == nil {
					log.Printf("[ERR] git: That repository does not exist: identifier: '%s', namespace: %+v", identifier, namespace)
					return false, nil
				}

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

// Package docker is a helper utility for performing the required docker
// functions and manipulations that Orbit requires.
package docker

import (
	"context"
	"log"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

// ForceUpdateService will forcefully update a service that has the ID
// specified.
func ForceUpdateService(id string) error {
	ctx := context.Background()
	cli, err := client.NewEnvClient()
	if err != nil {
		log.Printf("[ERR] docker: Could not create client: %s", err)
		return err
	}
	service, _, err := cli.ServiceInspectWithRaw(ctx, id)
	if err != nil {
		log.Printf("[ERR] docker: Could not access raw service inspection: %s", err)
		return err
	}
	service.Spec.TaskTemplate.ForceUpdate = 1 // Force update
	_, err = cli.ServiceUpdate(ctx, id, service.Meta.Version, service.Spec, types.ServiceUpdateOptions{})
	if err != nil {
		log.Printf("[ERR] docker: Could not update service: %s", err)
		return err
	}
	return nil
}

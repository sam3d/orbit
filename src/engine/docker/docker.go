// Package docker is a helper utility for performing the required docker
// functions and manipulations that Orbit requires.
package docker

import (
	"context"
	"log"
	"os"
	"os/exec"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

// ForceUpdateService will use the docker CLI directly to forcefully update a
// service with the given ID.
func ForceUpdateService(id string) error {
	cmd := exec.Command("docker", "service", "update", id, "--force")
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		log.Printf("[ERR] docker: Could not run service update%s", err)
		return err
	}
	log.Printf("[INFO] docker: Updated service %s", id)
	return nil
}

// forceUpdateService will forcefully update a service that has the ID
// specified.
func forceUpdateService(id string) error {
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
	service.Spec.TaskTemplate.ForceUpdate++ // Perform the force update
	_, err = cli.ServiceUpdate(ctx, id, service.Meta.Version, service.Spec, types.ServiceUpdateOptions{})
	if err != nil {
		log.Printf("[ERR] docker: Could not update service: %s", err)
		return err
	}
	return nil
}

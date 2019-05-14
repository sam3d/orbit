// Package docker is a helper utility for performing the required docker
// functions and manipulations that Orbit requires.
package docker

import (
	"context"
	"log"
	"net"
	"os"
	"os/exec"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

// SwarmInit will start a new instance of Docker Swarm with the specified ip
// address as the advertise address.
func SwarmInit(ip net.IP) error {
	ipStr := ip.String()
	cmd := exec.Command("docker", "swarm", "init", "--advertise-addr", ipStr)
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		log.Printf("[ERR] docker: Could not run swarm init: %s", err)
		return err
	}
	log.Printf("[INFO] docker: Swarm init with advertise address of %s", ipStr)
	return nil
}

// ForceLeaveSwarm will ensure that a node is not a member of a swarm before
// starting another one.
func ForceLeaveSwarm() error {
	cmd := exec.Command("docker", "swarm", "leave", "--force")
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		log.Printf("[ERR] docker: Could not run swarm leave with force: %s", err)
		return err
	}
	log.Printf("[INFO] docker: Force left existing swarm")
	return nil
}

// SwarmToken returns the specified token for connecting to this docker swarm
// instance. If the manager parameter is false it will be a worker token, if it
// is true it will be a manager token.
func SwarmToken(manager bool) string {
	var tokenType string
	if manager {
		tokenType = "manager"
	} else {
		tokenType = "worker"
	}

	cmd := exec.Command("docker", "swarm", "join-token", tokenType, "-q")
	output, err := cmd.Output()
	if err != nil {
		log.Printf("[ERR] docker: Could not retrieve output for %s token for swarm: %s", tokenType, err)
		return ""
	}

	return strings.TrimSpace(string(output))
}

// JoinSwarm will attempt to join the swarm with the given IP address and token.
func JoinSwarm(ip, token string) error {
	cmd := exec.Command("docker", "swarm", "join", "--token", token, ip)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Printf("{ERR] docker: Could not join swarm with token: %s", err)
		return err
	}
	return nil
}

// CreateOverlayNetwork will create a docker swarm network for overlay routing.
// This should be done after the swarm has been initialised, and only needs to
// be performed once per cluster.
func CreateOverlayNetwork(name string) error {
	cmd := exec.Command("docker", "network", "create", "-d", "overlay", name)
	if err := cmd.Run(); err != nil {
		log.Printf("[ERR] docker: Could not create overlay network with name %s: %s", name, err)
		return err
	}
	return nil
}

// ForceUpdateService will use the docker CLI directly to forcefully update a
// service with the given ID.
func ForceUpdateService(id string) error {
	cmd := exec.Command("docker", "service", "update", id, "--force")
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		log.Printf("[ERR] docker: Could not run service update: %s", err)
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

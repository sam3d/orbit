// Package docker is a helper utility for performing the required docker
// functions and manipulations that Orbit requires.
package docker

import (
	"bufio"
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
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

// DeployRegistry will create a docker service for the docker registry. It will
// use a local volume (as given by the data path) and make its available on the
// swarm nodes with the given port. This registry is where the built images are
// pushed so that they can be used on all nodes.
func DeployRegistry(path string, port int) error {
	cmd := exec.Command("docker", "service", "create",
		"--name", "registry",
		"--mount", fmt.Sprintf("type=bind,source=%s,target=/var/lib/registry", path),
		"--replicas", "1",
		"--publish", fmt.Sprintf("%d:5000", port),
		"registry:2",
	)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	log.Printf("[INFO] docker: Running command '%s'", strings.Join(cmd.Args, " "))
	if err := cmd.Run(); err != nil {
		log.Printf("[ERR] docker: Could not deploy registry server: %s", err)
		return err
	}

	return nil
}

// ServiceMount is a mount that a docker service uses.
type ServiceMount struct {
	Source string
	Target string
	Type   string
}

func (m ServiceMount) String() string {
	// Default to using a bind mount.
	if m.Type == "" {
		m.Type = "bind"
	}

	return fmt.Sprintf("type=%s,source=%s,target=%s", m.Type, m.Source, m.Target)
}

// EnsureDockerfile creates a herokuish dockerfile if one isn't present in the
// provided directory.
func EnsureDockerfile(path string) error {
	dockerfilePath := filepath.Join(path, "Dockerfile")
	if _, err := os.Stat(dockerfilePath); !os.IsNotExist(err) {
		// The file exists, so we don't need to do anything.
		return nil
	}

	// The file doesn't exist, so write the standard herokuish dockerfile.
	dockerfile := `FROM gliderlabs/herokuish
WORKDIR /tmp/build
COPY . .
RUN /build
`
	ioutil.WriteFile(dockerfilePath, []byte(dockerfile), 0644)

	return nil
}

// Build will perform the build process and output a channel containing the
// streaming build process that is taking place.
func Build(path, tag string) (<-chan string, <-chan error) {
	// Construct the output channel.
	outputCh := make(chan string)
	errorCh := make(chan error)

	go func() {
		// Ensure that on completion (or return) that the channel is closed.
		defer close(outputCh)
		defer close(errorCh)

		// Construct the build command.
		cmd := exec.Command("docker", "build", "-t", tag, path)
		rc, err := cmd.StdoutPipe()
		if err != nil {
			errorCh <- fmt.Errorf("could not pipe stdout from the docker build command: %s", err)
			return
		}

		// Handle the output from this command by the individual lines.
		scanner := bufio.NewScanner(rc)
		go func() {
			for scanner.Scan() {
				outputCh <- scanner.Text()
			}
		}()

		// Now we actually need to start the command.
		if err := cmd.Start(); err != nil {
			errorCh <- fmt.Errorf("could not start the command: %s", err)
			return
		}

		// Wait for the command to complete.
		if err := cmd.Wait(); err != nil {
			errorCh <- fmt.Errorf("could not wait for command: %s", err)
			return
		}
	}()

	// Provide the output channel to the caller.
	return outputCh, errorCh
}

// Publish is a port combination for a docker service.
type Publish struct {
	Host      int
	Container int
}

func (p Publish) String() string {
	return fmt.Sprintf("%d:%d", p.Host, p.Container)
}

// Service is a logical docker service. This is not a complete service
// description, but includes enough of the configuration for Orbit to function
// properly.
type Service struct {
	Name                 string
	Tag                  string
	Replicas             int
	DisableLocalRegistry bool
	Publish              []Publish
	Mode                 ServiceMode
	Mounts               []ServiceMount
	Networks             []string
}

// ServiceMode is a way in which to deploy a service.
type ServiceMode int

const (
	// Replicated means that a service gets replicated as specified.
	Replicated ServiceMode = iota
	// Global means that a service runs on each node that specifies it.
	Global
)

// CreateService will take in a service configuration object and create the
// docker service based on that.
func CreateService(services ...Service) error {
	for _, s := range services {
		if err := createService(s); err != nil {
			return err
		}
	}
	return nil
}

// createService performs the create service operation on a single service.
func createService(s Service) error {
	// Construct the basic arguments required for the service create command.
	args := []string{"service", "create"}

	// Add the name.
	if s.Name != "" {
		args = append(args, "--name", s.Name)
	}

	// Ensure there's at least one replica if that's the mode that ends up being
	// used.
	if s.Replicas == 0 {
		s.Replicas = 1
	}

	// Either set replicas or global mode.
	switch s.Mode {
	case Replicated:
		args = append(args, "--replicas", strconv.Itoa(s.Replicas))
	case Global:
		args = append(args, "--mode", "global")
	}

	// Add the mount declarations.
	for _, m := range s.Mounts {
		args = append(args, "--mount", m.String())
	}

	// Add the networks (and include orbit automatically).
	args = append(args, "--network", "orbit")
	for _, n := range s.Networks {
		args = append(args, "--network", n)
	}

	// Add the port bindings.
	for _, p := range s.Publish {
		args = append(args, "--publish", p.String())
	}

	// And finally, add the image tag. This can change depending upon whether the
	// service supplied includes a different registry specification to pull the
	// image from.
	if s.DisableLocalRegistry {
		args = append(args, s.Tag)
	} else {
		args = append(args, fmt.Sprintf("127.0.0.1:6510/%s", s.Tag))
	}

	cmd := exec.Command("docker", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	log.Printf("[INFO] docker: Running command '%s'", strings.Join(args, " "))
	if err := cmd.Run(); err != nil {
		log.Printf("[ERR] docker: Could not run docker service create on service %s: %s", s.Tag, err)
		return err
	}
	return nil
}

// Push will perform a docker push operation on a series of registry tags.
func Push(tags ...string) error {
	for _, tag := range tags {
		name := fmt.Sprintf("127.0.0.1:6510/%s", tag)
		cmd := exec.Command("docker", "push", name)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			log.Printf("[ERR] docker: Could not push image with tag %s: %s", tag, err)
			return err
		}
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

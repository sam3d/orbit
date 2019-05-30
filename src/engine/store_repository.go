package engine

import (
	"crypto/rand"
	"encoding/hex"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Repository is the information about the code that someone has on the system.
// This is generally a git repository, and it is used by the API git component
// to decide where to put the code.
type Repository struct {
	ID   string `json:"id"`
	Name string `json:"name"`

	NamespaceID string `json:"namespace_id"`
}

// gitBranches will list the branches in a given git repository.
func gitBranches(path string) (branches []string) {
	cmd := exec.Command("git", "branch")
	cmd.Dir = path
	output, err := cmd.Output()
	if err != nil {
		return branches
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		trim := strings.Trim(line, "* ")
		if trim != "" {
			branches = append(branches, trim)
		}
	}

	return branches
}

// gitCheckout will checkout a specified branch in a given path.
func gitCheckout(path, branch string) error {
	cmd := exec.Command("git", "checkout", branch)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

// Get list of files for a given git branch in a given bare path.
func gitFiles(path, branch string) (files []string) {
	cmd := exec.Command("git", "ls-tree", "--full-tree", "--name-only", "-r", branch)
	cmd.Dir = path
	output, err := cmd.Output()
	if err != nil {
		return files
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		file := strings.TrimSpace(line)
		if file == "" {
			continue
		}
		files = append(files, file)
	}

	return files
}

// RepoFiles will list all of the files with a given repository ID. You, the
// caller, must ensure that the repo directory that this will be searching
// already exists.
func (s *Store) RepoFiles(id string) (files map[string][]string) {
	files = make(map[string][]string)

	// Derive the repository location.
	volume := s.OrbitSystemVolume()
	if volume == nil {
		return
	}
	repoPath := filepath.Join(volume.Paths().Data, "repositories", id)

	// Now get a list of the git branches for the repository, and for each branch,
	// append a list of the files in that branch to the map.
	for _, branch := range gitBranches(repoPath) {
		files[branch] = gitFiles(repoPath, branch)
	}

	return files
}

// Repositories is a group of the store repository. The group allows for easier
// ID generation.
type Repositories []Repository

// GenerateID will create a unique identifier for the repository.
func (r *Repositories) GenerateID() string {
search:
	for {
		b := make([]byte, 8)
		rand.Read(b)
		id := hex.EncodeToString(b)

		for _, repo := range *r {
			if repo.ID == id {
				continue search
			}
		}

		return id
	}
}

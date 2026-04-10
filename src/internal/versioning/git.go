package versioning

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

var ErrNotGitRepo = errors.New("not a git repository")

type GitClient interface {
	IsRepo(root string) bool
	FetchTags(root string) error
	Tags(root string) ([]string, error)
	CreateTag(root, tag string) error
	StageFiles(root string, files []string) error
	Commit(root, message string) error
	Push(root string) error
	PushTag(root, tag string) error
}

type RealGit struct{}

func (RealGit) IsRepo(root string) bool {
	cmd := exec.Command("git", "rev-parse", "--is-inside-work-tree")
	cmd.Dir = root
	return cmd.Run() == nil
}

func (RealGit) FetchTags(root string) error {
	cmd := exec.Command("git", "fetch", "--tags")
	cmd.Dir = root
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git fetch --tags failed: %s", bytes.TrimSpace(output))
	}

	return nil
}

func (RealGit) Tags(root string) ([]string, error) {
	cmd := exec.Command("git", "tag")
	cmd.Dir = root
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("git tag failed: %s", bytes.TrimSpace(output))
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) == 1 && lines[0] == "" {
		return nil, nil
	}

	return lines, nil
}

func (RealGit) CreateTag(root, tag string) error {
	cmd := exec.Command("git", "tag", tag)
	cmd.Dir = root
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git tag %s failed: %s", tag, bytes.TrimSpace(output))
	}

	return nil
}

func (RealGit) StageFiles(root string, files []string) error {
	args := append([]string{"add", "--"}, files...)
	cmd := exec.Command("git", args...)
	cmd.Dir = root
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git add failed: %s", bytes.TrimSpace(output))
	}

	return nil
}

func (RealGit) Commit(root, message string) error {
	cmd := exec.Command("git", "commit", "-m", message)
	cmd.Dir = root
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git commit failed: %s", bytes.TrimSpace(output))
	}

	return nil
}

func (RealGit) Push(root string) error {
	cmd := exec.Command("git", "push")
	cmd.Dir = root
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git push failed: %s", bytes.TrimSpace(output))
	}

	return nil
}

func (RealGit) PushTag(root, tag string) error {
	cmd := exec.Command("git", "push", "origin", tag)
	cmd.Dir = root
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git push origin %s failed: %s", tag, bytes.TrimSpace(output))
	}

	return nil
}

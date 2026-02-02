package git

import (
	"errors"
	"os/exec"
	"regexp"
	"strings"
)

// RepoInfo holds GitHub repository information
type RepoInfo struct {
	Org  string
	Repo string
}

// GetRepoInfo extracts org/repo from the current git repository
func GetRepoInfo() (*RepoInfo, error) {
	cmd := exec.Command("git", "config", "--get", "remote.origin.url")
	output, err := cmd.Output()
	if err != nil {
		return nil, errors.New("not in a git repository with origin remote")
	}

	remoteURL := strings.TrimSpace(string(output))
	return ParseRemoteURL(remoteURL)
}

// ParseRemoteURL extracts org/repo from a GitHub remote URL
func ParseRemoteURL(remoteURL string) (*RepoInfo, error) {
	// Match various URL formats:
	// git@github.com:org/repo.git
	// https://github.com/org/repo.git
	// http://proxy@host/git/org/repo
	// http://host/git/org/repo

	// Try standard GitHub URL format first
	re := regexp.MustCompile(`github\.com[:/]([^/]+)/([^/.]+)(\.git)?$`)
	matches := re.FindStringSubmatch(remoteURL)
	if matches != nil {
		return &RepoInfo{
			Org:  matches[1],
			Repo: matches[2],
		}, nil
	}

	// Try proxy/mirror format: /git/org/repo
	re2 := regexp.MustCompile(`/git/([^/]+)/([^/.]+)(?:\.git)?$`)
	matches = re2.FindStringSubmatch(remoteURL)
	if matches != nil {
		return &RepoInfo{
			Org:  matches[1],
			Repo: matches[2],
		}, nil
	}

	// Try generic format: last two path components
	re3 := regexp.MustCompile(`[/:]([^/]+)/([^/.]+?)(?:\.git)?$`)
	matches = re3.FindStringSubmatch(remoteURL)
	if matches != nil {
		return &RepoInfo{
			Org:  matches[1],
			Repo: matches[2],
		}, nil
	}

	return nil, errors.New("could not parse GitHub org/repo from remote URL: " + remoteURL)
}

// Fetch fetches from origin
func Fetch() error {
	cmd := exec.Command("git", "fetch", "origin")
	cmd.Stdout = nil
	cmd.Stderr = nil
	return cmd.Run()
}

// FetchPrune fetches from origin with pruning
func FetchPrune() error {
	cmd := exec.Command("git", "fetch", "origin", "--prune")
	return cmd.Run()
}

// GetMainWorktreePath returns the path to the main worktree
func GetMainWorktreePath() (string, error) {
	cmd := exec.Command("git", "worktree", "list", "--porcelain")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "worktree ") {
			return strings.TrimPrefix(line, "worktree "), nil
		}
	}

	return "", errors.New("could not find main worktree")
}

// BranchExists checks if a branch exists locally
func BranchExists(branchName string) bool {
	cmd := exec.Command("git", "show-ref", "--verify", "--quiet", "refs/heads/"+branchName)
	return cmd.Run() == nil
}

// RemoteBranchExists checks if a branch exists on origin
func RemoteBranchExists(branchName string) bool {
	cmd := exec.Command("git", "show-ref", "--verify", "--quiet", "refs/remotes/origin/"+branchName)
	return cmd.Run() == nil
}

// DeleteBranch deletes a local branch
func DeleteBranch(branchName string) error {
	cmd := exec.Command("git", "branch", "-D", branchName)
	return cmd.Run()
}

// DeleteRemoteBranch deletes a branch from origin
func DeleteRemoteBranch(branchName string) error {
	cmd := exec.Command("git", "push", "origin", "--delete", branchName)
	return cmd.Run()
}

// GetLastCommitMessage returns the last commit message for the current branch or a specific ref
func GetLastCommitMessage(ref string) (string, error) {
	args := []string{"log", "-1", "--pretty=%B"}
	if ref != "" {
		args = append(args, ref)
	}
	cmd := exec.Command("git", args...)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

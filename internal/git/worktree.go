package git

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// HasUncommittedChanges checks if a directory has uncommitted git changes
func HasUncommittedChanges(path string) bool {
	cmd := exec.Command("git", "status", "--porcelain")
	cmd.Dir = path
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	return len(strings.TrimSpace(string(output))) > 0
}

// GetStatusShort returns the short git status output
func GetStatusShort(path string) (string, error) {
	cmd := exec.Command("git", "status", "--short")
	cmd.Dir = path
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// IsInsideWorktree checks if the current directory is inside the given worktree path
func IsInsideWorktree(worktreePath string) bool {
	cwd, err := os.Getwd()
	if err != nil {
		return false
	}
	return cwd == worktreePath || strings.HasPrefix(cwd, worktreePath+string(os.PathSeparator))
}

// FindWorktreeByIssue finds a worktree directory by issue number
func FindWorktreeByIssue(base string, issueNumber int) string {
	pattern := filepath.Join(base, strconv.Itoa(issueNumber)+"-*")
	matches, err := filepath.Glob(pattern)
	if err != nil || len(matches) == 0 {
		return ""
	}
	return matches[0]
}

// ListWorktrees returns all worktree directories for the given base path
func ListWorktrees(base string) ([]string, error) {
	if _, err := os.Stat(base); os.IsNotExist(err) {
		return nil, nil
	}

	entries, err := os.ReadDir(base)
	if err != nil {
		return nil, err
	}

	var worktrees []string
	for _, entry := range entries {
		if entry.IsDir() {
			worktrees = append(worktrees, filepath.Join(base, entry.Name()))
		}
	}
	return worktrees, nil
}

// DetectIssueNumber extracts issue number from current directory if inside a worktree
func DetectIssueNumber(base string) (int, bool) {
	cwd, err := os.Getwd()
	if err != nil {
		return 0, false
	}

	if !strings.HasPrefix(cwd, base) {
		return 0, false
	}

	// Get the relative path and extract the first directory component
	rel, err := filepath.Rel(base, cwd)
	if err != nil {
		return 0, false
	}

	// Get the first path component (the worktree directory name)
	parts := strings.Split(rel, string(os.PathSeparator))
	if len(parts) == 0 {
		return 0, false
	}
	dirname := parts[0]

	// Extract issue number from directory name (e.g., "42-fix-bug")
	re := regexp.MustCompile(`^(\d+)-`)
	matches := re.FindStringSubmatch(dirname)
	if matches == nil {
		return 0, false
	}

	num, err := strconv.Atoi(matches[1])
	if err != nil {
		return 0, false
	}

	return num, true
}

// CreateWorktree creates a new git worktree
func CreateWorktree(path, branchName, baseBranch string) error {
	// Ensure parent directory exists
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	cmd := exec.Command("git", "worktree", "add", path, "-b", branchName, baseBranch)
	cmd.Stdout = os.Stderr // Output to stderr so it doesn't interfere with path capture
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// CreateWorktreeFromBranch creates a worktree from an existing branch
func CreateWorktreeFromBranch(path, branchName string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	cmd := exec.Command("git", "worktree", "add", path, branchName)
	cmd.Stdout = os.Stderr // Output to stderr so it doesn't interfere with path capture
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// CreateWorktreeFromRemote creates a worktree tracking a remote branch
func CreateWorktreeFromRemote(path, branchName, remoteBranch string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	cmd := exec.Command("git", "worktree", "add", path, "-b", branchName, remoteBranch)
	cmd.Stdout = os.Stderr // Output to stderr so it doesn't interfere with path capture
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// RemoveWorktree removes a git worktree
func RemoveWorktree(path string, force bool) error {
	args := []string{"worktree", "remove", path}
	if force {
		args = append(args, "--force")
	}
	cmd := exec.Command("git", args...)
	err := cmd.Run()

	// If the worktree removal failed (e.g., exit status 128), check if the directory still exists
	if err != nil {
		if _, statErr := os.Stat(path); statErr == nil {
			// Directory exists but git doesn't know about it anymore
			// This happens when git's worktree tracking is broken
			// Try to remove the directory manually
			removeErr := os.RemoveAll(path)
			if removeErr != nil {
				// If we can't remove it, return both errors
				return fmt.Errorf("git worktree remove failed: %v, manual removal failed: %v", err, removeErr)
			}
			// Successfully removed the orphaned directory
			return nil
		}
	}

	return err
}

// PruneWorktrees prunes worktrees that no longer exist
func PruneWorktrees() (string, error) {
	cmd := exec.Command("git", "worktree", "prune", "-v")
	output, err := cmd.CombinedOutput()
	return string(output), err
}

// Push pushes a branch to origin
func Push(path, branchName string) error {
	cmd := exec.Command("git", "push", "-u", "origin", branchName)
	cmd.Dir = path
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// GetAheadBehind returns the ahead/behind counts for a branch relative to its remote
func GetAheadBehind(path, branchName string) (ahead, behind int, err error) {
	cmd := exec.Command("git", "rev-list", "--left-right", "--count", "origin/"+branchName+"...HEAD")
	cmd.Dir = path
	output, err := cmd.Output()
	if err != nil {
		return 0, 0, err
	}

	parts := strings.Fields(strings.TrimSpace(string(output)))
	if len(parts) != 2 {
		return 0, 0, nil
	}

	behind, _ = strconv.Atoi(parts[0])
	ahead, _ = strconv.Atoi(parts[1])
	return ahead, behind, nil
}

// GetLocalBranches returns all local branch names
func GetLocalBranches() ([]string, error) {
	cmd := exec.Command("git", "branch", "--format=%(refname:short)")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var branches []string
	for _, line := range lines {
		if line != "" {
			branches = append(branches, line)
		}
	}
	return branches, nil
}

// GetUncommittedCount returns the number of uncommitted changes
func GetUncommittedCount(path string) int {
	cmd := exec.Command("git", "status", "--short")
	cmd.Dir = path
	output, err := cmd.Output()
	if err != nil {
		return 0
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) == 1 && lines[0] == "" {
		return 0
	}
	return len(lines)
}
